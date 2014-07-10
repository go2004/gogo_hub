/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-2-25
 * Time: 下午4:40
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"net"
	//	"log"
		"packet"
	//	"os"
	"sync"
	"time"
	//"fmt"
	"common"
	"runtime"
	//	"helper"
)

//// 命令接口
//type CmdInterface interface {
//	doMsg(pUser *user, length int, b []byte) bool
//}

//全局变量
var (
	allSocketsSem    sync.RWMutex       // Used to synchronize access to all data structures in this var group
	allConn       [Max_Conn]*connection // This array contains all socket
	//lastPlayerSlot   int                      // The last slot in use in allPlayers
	//numPlayers       int                      // The total number of players currently
	//CmdList        [Max_CMD]CmdInterface //命令最大数量
	//ProtoHandler [Max_CMD]func ( *user, *packet.Packet) //协议涵数
)

//开始服务
func StartListen() error {
	addr := common.GetElement("SocketInfo", "ListenAddr", "127.0.0.1:8080")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	LogInfo("StartListen start, addr= ",addr);
	go func() {
		for failures := 0; failures < 100; {
			if !WorldRunning {
				break;
			}

			conn, err := listener.Accept()
			if err != nil {
				LogError("Failed listening: ", err, "\n");
				failures++
			}
			ok, freeSocket := GetFreeConnection(conn)
			if ok {
				go NewConnection(freeSocket)
			} else {
				LogError("GetFreeConnection > %d\n", Max_Conn);
			}
		}
		LogError("Too many listener.Accept() errors, giving up")
		//os.Exit(1)
	}()
	return nil
}

//初始化连接数据
func InitConnection() {
	var index uint32 = 0
	for i := 0; i < Max_Conn; i++ {
		up := new(connection)
		up.id = index
		up.conn = nil
		up.connState = ConnStateFree
		up.logonTimer = time.Now()
		up.channel = make(chan []byte, ClientChannelSize)
		up.commandChannel = make(chan ClientCommand, ClientChannelSize)
		up.user = nil
		allConn[i] = up
		index = index + 1
	}
}

//空闲连接
func GetFreeConnection(conn net.Conn) (ok bool, pConn *connection) {
	allSocketsSem.Lock()
	var index int
	for index = 0; index < Max_Conn; index++ {
		if allConn[index].connState == ConnStateFree {
			allConn[index].conn = conn
			allConn[index].connState = ConnStateLogin
			allConn[index].user = nil
			break
		}
	}
	allSocketsSem.Unlock()
	if index >= Max_Conn {
		LogInfo(" Handle the case with too many players,index = ", index, ",MAX_PLAYERS =\n", Max_Conn)
		return false, nil
	}
	return true, allConn[index]
}

//释放连连接
func FreeConnection(pConn *connection) (ok bool ) {
	//fmt.Printf(" FreeConnection id = %d\n",pSocket.id)
	allSocketsSem.Lock()
	pConn.connState = ConnStateFree;
	pConn.conn.Close();
	pConn.conn = nil;
	pConn.user = nil
	if len(pConn.channel) > 0 {
		tmp := <-pConn.channel
		LogInfo(" clear channel =", tmp)
	}

	if len(pConn.commandChannel) > 0 {
		tmp := <-pConn.commandChannel
		LogInfo(" clear channel =", tmp)
	}
	allSocketsSem.Unlock()
	return true
}

//关闭所有连接
func CloseConnection() {
	for i := 0; i < Max_Conn; i++ {
		pSocket := allConn[i]
		FreeConnection(pSocket)
	}
}

func NewConnection(pConn *connection) {
	defer catchError(pConn)

	var length int
	buff := pConn.recvBuff[:]

	tick := 0
	for tick = 0; tick < Max_Tick; {
		if !WorldRunning {
			break;
		}
		time.Sleep(ObjectsUpdatePeriod)
		// Read out all waiting data, if any, from the incoming channels
		for moreData := true; moreData; {
			select {
			case clientMessage := <-pConn.channel:
				pConn.writeBlocking_Bl(clientMessage)
			case clientCommand := <-pConn.commandChannel:
				clientCommand(pConn)
			default:
				moreData = false
			}
			if pConn.connState == ConnStateDisc {   //玩家关闭
				break;
			}
		}

		//Read data length from socket
		pConn.conn.SetReadDeadline(time.Now().Add(ObjectsUpdatePeriod))
		n, err := pConn.conn.Read(buff[0:2]) // Read the length information. This will block for ObjectsUpdatePeriod ns
		if err != nil {
			if e2, ok := err.(*net.OpError); ok && (e2.Timeout() || e2.Temporary()) {
				tick += 1; //log.Printf("Read timeout %v", e2) // This will happen frequently
				continue
			}
			LogInfo("Disconnect from ", pConn.id, ", because of ", err)        // This is a normal case

			break
		}
		//包没收完，继续收
		if n == 1 {
			//LogDebug("Got %d bytes reading from socket\n", n)
			for (n == 1) {
				n2, err := pConn.conn.Read(buff[1:2]) // Second byte of the length
				if err != nil || n2 != 1 {
					LogInfo("Failed again to read: ", err)
					break;
				}
				n = 2
			}
		}
		if n != 2 {
			LogDebug("Got %d bytes reading from socket,tick =%d\n", n, tick)
			continue
		}
		//ret = uint16(buf[0])<<8 | uint16(buf[1])
		length = int(uint(buff[0])<<8 + uint(buff[1]))
		if length < 4 || length > Max_Recv_Packge {        //超过最大包长，异常，直接丢掉
			LogInfo("Expecting ", length, " bytes, which is too much for buffer. Buffer was extended")
			continue
		}

		// Read the rest of the bytes
		pConn.conn.SetReadDeadline(time.Now().Add(ObjectsUpdatePeriod)) // Give it up after a long time
		for n2 := n; n2 < length; {
			//LogDebug("Got", n2, "out of", length, "bytes")
			// If unlucky, we only get one byte at a time
			const maxDelay = 5
			for i := 0; i < maxDelay; i++ {
				// Normal case is one iteration in this loop
				n, err = pConn.conn.Read(buff[n2:length])
				if err == nil {
					break // No error
				}
				e2, ok := err.(net.Error)
				if ok && !e2.Temporary() && !e2.Timeout() { // We don't really expect a timeout here
					break // Bad error, can't handle it.
				}

				LogInfo("Temporary, retry")

				if i == maxDelay - 1 {
					LogInfo("Timeout, giving up")
				}
			}
			if err != nil {
				LogInfo("Disconnect ", pConn.id, " because of ", err)
				panic(err)
			}
			n2 += n
		}

		//加入到流量统计中
		//trafficStatistics.AddReceived(length)
		//cmd := uint16(buff[2])<<8 | uint16(buff[3])
		reader := packet.Reader(buff[2:length])
		cmd,err := reader.ReadU16()

		//登录验证包及前置验证
		if (pConn.connState <= ConnStateLogin) {
			PrepareHandler(pConn, cmd, reader)
			tick += 1;
			continue;
		}

		//正常数据
		if pConn.user != nil && pConn.user.conn != nil {
			ProtoHandler(pConn, cmd, reader)
			tick = 0
			continue;
		}

		//异常数据，TICK计次
		tick += 1;
		LogDebug("doMsg empty,Cmd = ", cmd)
	}
	LogInfo("Disconnect ", pConn.id, " tick = ", tick, "(", Max_Tick, ").\n")

}

//异常处理
func catchError(pConn *connection) {
	if err := recover(); err != nil {
		LogInfo("err", err)

		//这里要异常要退出整个程序，只能在debug模式下使用
		for i := 0; i < 10; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if ok {
				LogInfo("frame ", i, ":[func:", runtime.FuncForPC(funcName).Name(), ",file:", file, ",line:", line)
			}
		}
	}

	if pConn == nil {return}

	LogInfo("Disconnect id = ", pConn.id)
	FreeConnection(pConn)
}



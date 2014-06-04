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
	"logger"
	//	"io"
	//	"os"
	"sync"
	"time"
	//	"log"
	"common"
)

// 命令接口
type CmdInterface interface {
	doMsg(pSocket *socket, length int, b []byte) bool
}

//全局变量
var (
	allSocketsSem    sync.RWMutex        // Used to synchronize access to all data structures in this var group
	allSocket       [MAX_PLAYERS]*socket // This array contains all players
	//lastPlayerSlot   int                      // The last slot in use in allPlayers
	//numPlayers       int                      // The total number of players currently
	CmdList        [MAX_CMD]CmdInterface //命令最大数量
)

//开始服务
func StartListen() error {
	addr := common.GetElement("SocketInfo", "ListenAddr", "0.0.0.0:8080")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		for failures := 0; failures < 100; {
			if !WorldRunning {
				break;
			}

			conn, err := listener.Accept()
			if err != nil {
				logger.Error("Failed listening: ", err, "\n");
				failures++
			}
			ok, freeSocket := GetFreeConnection(conn)
			if ok {
				go NewClient(freeSocket)
			} else {
				logger.Error("GetFreeConnection > %d\n", MAX_PLAYERS);
			}
		}
		logger.Error("Too many listener.Accept() errors, giving up")
		//os.Exit(1)
	}()
	return nil
}

//初始化连接数据
func InitConnection() {
	var index uint32
	index = 0
	for i := 0; i < MAX_PLAYERS; i++ {
		up := new(socket)
		up.id = index
		up.conn = nil
		up.connState = ConnStateFree
		up.logonTimer = time.Now()
		up.channel = make(chan []byte, ClientChannelSize)
		up.commandChannel = make(chan ClientCommand, ClientChannelSize)
		allSocket[i] = up
		index = index + 1
	}
}

//空闲连接
func GetFreeConnection(conn net.Conn) (ok bool, freesocket *socket) {
	allSocketsSem.Lock()
	var index int
	for index = 0; index < MAX_PLAYERS; index++ {
		if allSocket[index].connState == ConnStateFree {
			allSocket[index].conn = conn
			allSocket[index].connState = ConnStateLogin
			break
		}
	}
	allSocketsSem.Unlock()
	if index >= MAX_PLAYERS {
		logger.Info(" Handle the case with too many players,index = ", index, ",MAX_PLAYERS =\n", MAX_PLAYERS)
		return false, nil
	}


	return true, allSocket[index]
}

//释放连连接
func FreeConnection(pSocket *socket) (ok bool ) {
	//fmt.Printf(" FreeConnection id = %d\n",pSocket.id)
	allSocketsSem.Lock()
	pSocket.connState = ConnStateFree;
	pSocket.conn.Close();
	pSocket.conn = nil;
	if len(pSocket.channel) > 0 {
		tmp := <-pSocket.channel
		logger.Info(" clear channel =", tmp)
	}

	if len(pSocket.commandChannel) > 0 {
		tmp := <-pSocket.commandChannel
		logger.Info(" clear channel =", tmp)
	}
	allSocketsSem.Unlock()
	return true
}

//关闭所有连接
func CloseConnection() {
	for i := 0; i < MAX_PLAYERS; i++ {
		pSocket := allSocket[i]
		FreeConnection(pSocket)
	}
}

func NewClient(pSocket *socket) {
	//buff := make([]byte, Max_Recv_Packge) // Command buffer, also used for blocking messages.
	defer catchError(pSocket)
	var (
		buff [Max_Recv_Packge]byte
		cmd uint16
		length int
	);

	tick := 0
	for tick = 0; tick < Max_Tick; {
		if !WorldRunning {
			break;
		}

		// Read out all waiting data, if any, from the incoming channels
		for moreData := true; moreData; {
			select {
			case clientMessage := <-pSocket.channel:
				pSocket.writeBlocking_Bl(clientMessage)
			case clientCommand := <-pSocket.commandChannel:
				clientCommand(pSocket)
			default:
				moreData = false
			}
			if pSocket.connState == ConnStateDisc {   //玩家关闭
				break;
			}
		}

		//Read data length from socket
		pSocket.conn.SetReadDeadline(time.Now().Add(ObjectsUpdatePeriod))
		n, err := pSocket.conn.Read(buff[0:2]) // Read the length information. This will block for ObjectsUpdatePeriod ns
		if err != nil {
			if e2, ok := err.(*net.OpError); ok && (e2.Timeout() || e2.Temporary()) {
				tick += 1; //log.Printf("Read timeout %v", e2) // This will happen frequently
				continue
			}
			logger.Info("Disconnect from ", pSocket.id, ", because of ", err)        // This is a normal case

			break
		}
		//包没收完，继续收
		if n == 1 {
			//logger.Debug("Got %d bytes reading from socket\n", n)
			for (n == 1) {
				n2, err := pSocket.conn.Read(buff[1:2]) // Second byte of the length
				if err != nil || n2 != 1 {
					logger.Info("Failed again to read: ", err)
					break;
				}
				n = 2
			}
		}
		if n != 2 {
			logger.Debug("Got %d bytes reading from socket,tick =%d\n", n, tick)
			continue
		}

		length = int(uint(buff[1])<<8 + uint(buff[0]))
		if length > Max_Recv_Packge {        //超过最大包长，异常，直接丢掉
			logger.Info("Expecting ", length, " bytes, which is too much for buffer. Buffer was extended")
			continue
		}

		// Read the rest of the bytes
		pSocket.conn.SetReadDeadline(time.Now().Add(ObjectsUpdatePeriod)) // Give it up after a long time
		for n2 := n; n2 < length; {
			//logger.Debug("Got", n2, "out of", length, "bytes")
			// If unlucky, we only get one byte at a time
			var e2 net.Error
			const maxDelay = 5
			for i := 0; i < maxDelay; i++ {
				// Normal case is one iteration in this loop
				n, err = pSocket.conn.Read(buff[n2:length])
				if err == nil {
					break // No error
				}
				e2, ok := err.(net.Error)
				if ok && !e2.Temporary() && !e2.Timeout() { // We don't really expect a timeout here
					break // Bad error, can't handle it.
				}

				logger.Info("Temporary, retry")

				if i == maxDelay - 1 {
					logger.Info("Timeout, giving up")
				}
			}
			if err != nil {
				logger.Info("Disconnect ", pSocket.id, " because of ", err)
				if e2 != nil {
					logger.Info("Temporary: ", e2.Temporary(), ", Timeout: ", e2.Timeout())
				}

				panic(err)
			}
			n2 += n
		}

		if (length < 4) {    //包长度不正确，异常
			logger.Debug("Cmd = ", cmd, " length =(", length, ") <4")
			continue
		}

		//加入到流量统计中
		//trafficStatistics.AddReceived(length)

		cmd = uint16(uint(buff[3])<<8 + uint(buff[2]))
		cmdInterface := CmdList[cmd]
		bodyLen := length - 4
		if cmdInterface != nil {
			cmdInterface.doMsg(pSocket, bodyLen, ([]byte)(buff[4:length]))
			tick = 0;
		}else {
			tick += 1;
			logger.Debug("doMsg empty,Cmd = ", cmd)
			//fmt.Println("doMsg empty,Cmd = ",cmd)
		}


	}
	logger.Error("Disconnect %d tick = %d(%d).\n", pSocket.id, tick, Max_Tick)
	//End:
	//	log.Printf("Disconnect id = %d \n", pSocket.id)
	//	FreeConnection(pSocket)

}

//异常处理
func catchError(pSocket *socket) {
	if err := recover(); err != nil {
		logger.Info("err", err)
	}
	if pSocket == nil {return}

	logger.Info("Disconnect id = ", pSocket.id)
	FreeConnection(pSocket)
}



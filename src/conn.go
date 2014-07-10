/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-4-9
 * Time: 下午4:12
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"net"
	"time"
	//	"fmt"
)

type ClientCommand func (pConn *connection)
type connection struct {
	id					   	    uint32      //
	conn                        net.Conn    // The TCP/IP connectin to the player.
	connState                   uint8       // The player connection state. See definition of PlayerConnState*
	channel                     chan []byte // Data to be sent to the client is only handled by the listener process, all else must go through this channel. See writeNonBlocking()
	logonTimer                  time.Time   // Used to keep track of how long he player has been online
	commandChannel              chan ClientCommand
	tick		          	   	uint //tick number
	user       					*user  // This array contains all socket
	recvBuff 					[Max_Recv_Packge]byte   // Data  from to the client
}

//socket连接状态
const (
	ConnStateFree  = iota // The player has to free
	ConnStateLogin = iota // The player has to login
	ConnStatePass  = iota // The player has to provide a password
	ConnStateIn    = iota // The player is in the world
	ConnStateDisc  = iota // The player is logged in but disconnected
)

func (pConn *connection) writeBlocking_Bl(b []byte) {
	if ((pConn.connState == ConnStateDisc) || (pConn.connState == ConnStateFree)) { // Connection no longer available, don't even try
		return
	}
	for len(b) > 0 {
		n, err := pConn.conn.Write(b)
		//		trafficStatistics.AddSend(len(b))

		if err == nil {
			b = b[n:]
		} else if e2, ok := err.(*net.OpError); ok && (e2.Temporary() || e2.Timeout()) {
			continue
		} else {
			LogError("writeBlocking_Bl ", err, "%#v\n", err, err)
			return
		}
	}
	return
}

// Send a message to a client, but it must not block. Because of that, send the message to the
// local client process that can handle the blocking.
// Condition: There is no guarantee to be any locks, which means there is no guarantee in what order
// these messages arrive to the client. Ok, the order will stay the same, but there may be other
// processes that manage to inject message in between others.
func (pConn *connection) writeNonBlocking(b []byte) {
	if length := int(b[1]) + int(b[0])<<8; length != len(b) {
		LogError("Wrong length of message ,length=", length, ",len=%d", len(b))
		return
	}
	select {
	case pConn.channel <- b:
	default:
	}
}

// Send a non blocking command to the player.
// The command can fail if the receiver is full.
func (pConn *connection) SendCommand(cmd ClientCommand) {
	// Use a select statement to make sure it never blocks.
	select {
	case pConn.commandChannel <- cmd:
	default:
	}
}

//
//func (pConn *connection) CmdLogin(buff []byte) {
//	loginResult := DbConn.CmdLoginDbCheck("abcd", "a")
//	fmt.Println("loginResult = ",loginResult)
//	//验证通过
//	if !loginResult {
//		pConn.writeBlocking_Bl(buff)
//		return
//	}
//
//	nResult, newUser := GetFreeUser()
//	if nResult {
//		SetConn2User(pConn, newUser)
//		pConn.connState = ConnStatePass
//	}
//
//	//返回登录结果
//	pConn.writeBlocking_Bl(buff)
//}

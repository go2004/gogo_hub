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
	"logger"
	"time"
)

type ClientCommand func (mysocket *socket)
type socket struct {
	id					   	uint32      //
	conn                         net.Conn    // The TCP/IP connectin to the player.
	connState                    uint8       // The player connection state. See definition of PlayerConnState*
	channel                      chan []byte // Data to be sent to the client is only handled by the listener process, all else must go through this channel. See writeNonBlocking()
	logonTimer                   time.Time   // Used to keep track of how long he player has been online
	commandChannel               chan ClientCommand
	tick		          	   uint //tick number
}

//socket连接状态
const (
	ConnStateFree  = iota // The player has to free
	ConnStateLogin = iota // The player has to login
	ConnStatePass  = iota // The player has to provide a password
	ConnStateIn    = iota // The player is in the world
	ConnStateDisc  = iota // The player is logged in but disconnected
)

func (socket *socket) writeBlocking_Bl(b []byte) {
	if ((socket.connState == ConnStateDisc) || (socket.connState == ConnStateFree)) { // Connection no longer available, don't even try
		return
	}
	for len(b) > 0 {
		n, err := socket.conn.Write(b)
		//		trafficStatistics.AddSend(len(b))
		if err == nil {
			b = b[n:]
		} else if e2, ok := err.(*net.OpError); ok && (e2.Temporary() || e2.Timeout()) {
			continue
		} else {
			logger.Error("writeBlocking_Bl %v %#v\n", err, err)
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
func (socket *socket) writeNonBlocking(b []byte) {
	if length := int(b[0]) + int(b[1])<<8; length != len(b) {
		panic("Wrong length of message")
	}
	select {
	case socket.channel <- b:
	default:
	}
}

// Send a non blocking command to the player.
// The command can fail if the receiver is full.
func (socket *socket) SendCommand(cmd ClientCommand) {
	// Use a select statement to make sure it never blocks.
	select {
	case socket.commandChannel <- cmd:
	default:
	}
}

/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-4-9
 * Time: 下午5:10
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	//	"log"
	//	"io"
	//"os"
	//	"sync"
	//	"time"
	"fmt"
)

//
//type I2 interface { Get() int64; Put(int64) }

// 测试(实例)
type testPacket struct {
	length int //消息长度
}

//测试的接收函数(实例)
func (r *testPacket) doMsg(pSocket *socket, bodylength int, buff []byte) bool {
	if (pSocket == nil || bodylength != len(buff)) {
		return false
	}
	//fmt.Println("cmd = ",length,buff)
	var tmpBuff[3] byte = [3]byte{3, 0, buff[0] }
	pSocket.writeNonBlocking(tmpBuff[:])
	fmt.Println("pSocket bodylength= ", bodylength, tmpBuff)
	return true
}

//接收指令注册接口
func RegistCmd() {
	CmdList[CMD_SAVE] = &testPacket{length: 1}
}

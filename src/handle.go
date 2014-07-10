package main

import (
	"packet"
)

//接收指令注册接口
func RegistCmd() {
	//ProtoHandler[CMD_LOGIN] = P_user_login_info,
	//ProtoHandler[CMD_HEART] = P_heart_beat_req
}

//前置验证接口
func PrepareHandler(pConn *connection, cmd uint16, reader *packet.Packet) {

	switch cmd {
	case CMD_LOGIN:P_user_login_info(pConn, reader)

	default:
		LogInfo("Unknown command '%d'.\n", cmd)
		return
	}
}

//所有命令总接口
func ProtoHandler(pConn *connection, cmd uint16,  reader *packet.Packet) {
	switch cmd {
	case CMD_PING:P_PingAct(pConn, reader)

	default:
		LogInfo("Unknown command cmd=", cmd)
		return
	}

}

//登录
func P_user_login_info(pConn *connection, reader *packet.Packet) {
	tbl, err := PKT_user_login_info(reader)
	if (err != nil){return}

	LogInfo("tbl.userId =", tbl.userId)

	loginResult := DbConn.CmdLoginDbCheck("gogo", "a")
	LogInfo("loginResult = ", loginResult, "pConn = ", pConn.connState)

	//返回登录结果
	write := packet.Writer(1)
	write.WriteBool(loginResult)
	write.Calculate()
	pConn.conn.Write(write.Data())

	//验证没通过
	if !loginResult {
		return
	}

	nResult, newUser := GetFreeUser()
	if nResult {
		SetConn2User(pConn, newUser)
		pConn.connState = ConnStatePass
	}
}

func P_PingAct(pConn *connection, reader *packet.Packet) {
	writer := packet.Writer(CMD_PING)
	writer.WriteU32(pConn.id+100)
	writer.Calculate()
	pConn.writeBlocking_Bl(writer.Data())
}


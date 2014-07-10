/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-2-25
 * Time: 下午3:06
 * 协义结构定义
 */

package main

import "packet"

type user_login_info struct {
	userId       uint8
	//	F_client_version int32
	//	F_new_user       bool
	//	F_user_name      string
}

func PKT_user_login_info(reader *packet.Packet) (tbl user_login_info, err error) {
	tbl.userId, err = reader.ReadByte()
	PKT_checkErr(err)

	return
}

func PKT_checkErr(err error) {
	if err != nil {
		LogDebug(err)
		//		panic("error occured in protocol module")
	}
}

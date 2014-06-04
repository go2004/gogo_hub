package main

//
// This is a general configuration file for the game engine
// Times are defined in nonoseconds. That means 1e8=0.1s, 1e9=1s, 1e10=10s, 1e11=100s, etc.
// TODO: All parameters here should have precix Cnfg.
//
const (
	// How many nanoseconds between update of player and monster positions
	MainFrameUpdate      = 100000000            // 100 times per second update     100000000  1e7
	ObjectsUpdatePeriod  = 1e8                  // 10 times per second update
	MAX_PLAYERS          = 2000                 // The total number of players currently
	ClientChannelSize    = 100                  // Number of messages that can wait for being sent
	MAX_CMD              = 1024                 //指令列表
	Max_Recv_Packge      = 128                  //接收包最大长度
	Max_Tick             = 1024                 //最大心跳次数，超过则关闭SOCKET
	ConfigFile           = "config/default.ini" //配置文件路径
	SqlChannelSize       = 100                  //sql语句缓存大小
	DbUpdateQuerychannel = 1e8                  //10 times per second update  querychannel
	DbUpdatePingMaxTick  = 1024                 //ping数据库的TICK上限

)

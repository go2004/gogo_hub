package main

import (
	"log"
	"common"
	"strconv"
	"os"
	"runtime/pprof"
	"fmt"
)

//
// This is a general configuration file for the game engine
// Times are defined in nonoseconds. That means 1e8=0.1s, 1e9=1s, 1e10=10s, 1e11=100s, etc.
// TODO: All parameters here should have precix Cnfg.
//
const (
	// How many nanoseconds between update of player and monster positions
	MainFrameUpdate      = 100000000            // 100 times per second update     100000000  1e7
	ObjectsUpdatePeriod  = 1e8                  // 10 times per second update
	Max_Conn             = 2000                 // The total number of TCP/IP connectin
	MAX_PLAYERS            = 2000                 // The total number of player currently
	ClientChannelSize    = 16384                // Number of messages that can wait for being sent  1024*16
	Max_CMD              = 1024                 //指令列表
	Max_Recv_Packge      = 128                  //接收包最大长度
	Max_Tick             = 1024                 //最大心跳次数，超过则关闭SOCKET
	ConfigFile           = "config/default.ini" //配置文件路径
	SqlChannelSize       = 100                  //sql语句缓存大小
	DbUpdateQuerychannel = 1e8                  //10 times per second update  querychannel
	DbUpdatePingMaxTick  = 1024                 //ping数据库的TICK上限
)

const (
	ALL uint8 = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

func SetLogLevel(level uint8) {
	LogLevel = level
}

//加载配置
func initConfig() {
	common.LoadConfig(ConfigFile)
	levelStr := common.GetElement("main", "LogLevel", "1");
	level, _ := strconv.Atoi(levelStr)
	newLevel := uint8(level)
	SetLogLevel(newLevel)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			LogError(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

//------------------------------------------------ 严重程度由高到低
func LogDebug(v ...interface{}) {
	if LogLevel > DEBUG { return }
	log.Printf("[DEBUG] %v", fmt.Sprintln(v...))
}

func LogInfo(v ...interface{}) {
	if LogLevel > INFO { return }
	log.Printf("[INFO] %v", fmt.Sprintln(v...))
}

func LogWarn(v ...interface{}) {
	if LogLevel > WARN { return }
	log.Printf("[WARN] %v", fmt.Sprintln(v...))
}

func LogError(v ...interface{}) {
	if LogLevel > ERROR { return }
	log.Printf("[ERROR] %v", fmt.Sprintln(v...))
}

func LogFatal(v ...interface{}) {
	if LogLevel > FATAL { return }
	log.Printf("[FATAL] %v", fmt.Sprintln(v...))
}

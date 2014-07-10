/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-2-25
 * Time: 下午3:06
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"runtime"
	"log"
)

var (
	//ipPort      = flag.String("i", "127.0.0.1:8080", "IP port to listen on")
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file")
	logFileName        = "worldserver"
	LogLevel           = DEBUG                  //日志等级(默认等级)
	GC_INTERVAL int64  = 300 					// voluntary GC interval

	DbConn *DbMysql
	WorldRunning = true
)

func main() {
	//开启日志功能
	logPathFile := fmt.Sprintf("log/%s_%s.log", logFileName,time.Now().Format("20140101150405"))
	logFile, err := os.OpenFile(logPathFile, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
	if (err != nil) {
		return
	}
	log.SetOutput(logFile)
	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	defer logFile.Close()

	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	initConfig()

	//var err error
	DbConn, err = DbInit();
	if (err != nil) {
		LogError("Start DbInit is failure ", err);
		os.Exit(1)
	}
	InitConnection();
	LogInfo("InitConnection succes.... ");

	//RegistCmd(); //注册指令集
	//LogInfo("RegistCmd succes.... ");

	InitGlobalUser()

	nResult := StartListen()
	if nResult != nil {        //启动服务错误
		LogError("StartListen is failure =\n", nResult);
		os.Exit(1)
	}
	LogInfo("StartListen succes.... ");

	//主线程更新
	go mainFrameUpdate();

	//等待系统指令
	SignalProc();
}

//主线程更新
func mainFrameUpdate() {
	start := time.Now()
	elapsed := time.Now().Sub(start)
	for {
		start = time.Now()
		//更新涵数


		elapsed = time.Now().Sub(start)
		if elapsed > MainFrameUpdate*2 {
			LogInfo("mainFrameUpdate is Timeout[", MainFrameUpdate, "], elapsed=", elapsed, ", DateTime =",time.Now())
		}

		//gc处理
		//helper.SysRoutine()
		time.Sleep((MainFrameUpdate))
		//有退出标记，退出主线程
		if !WorldRunning {
			break
		}
	}
	LogError("Start world server stop ......\n")
}

//正常退出
func ServerTerminate() {
	if !WorldRunning {
		WorldRunning = false;
		time.Sleep(1e10*60)	//1分钟后关闭

		//下面进行些关闭操作
		CloseConnection();
		//DbClose(DbConn);
	}
}



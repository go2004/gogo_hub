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
	//	"fmt"
	"os"
	"bufio"
	"time"
	"runtime"
	"runtime/pprof"
	//	"log"
	//	"timerstats"
	"common"
	"logger"
)

var (
	ipPort     = flag.String("i", "127.0.0.1:8080", "IP port to listen on")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	//	logOnStdout = flag.Bool("s", false, "Send log file to standard otput")
	//	verboseFlag = flag.Int("v", 0, "Verbose, Higher number gives more");

	logFileName        = "worldserver.log"
	logLevel           = logger.DEBUG
	maxFileCount int32 = 1024;
	maxFileSize  int64 = 1024*10;

	DbConn *DbMysql
	WorldRunning = true
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	initConfig()


	//	if !*logOnStdout {
	//		logFile, _ := os.OpenFile(*logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	//		log.SetOutput(logFile)
	//		defer logFile.Close()
	//	}
	//	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)


	//
	var err error
	//	DbConn, err = DbInit();
	DbInit();
	if (err != nil) {
		logger.Error("Start DbInit is failure ", err);
		os.Exit(1)
	}
	InitConnection();
	logger.Info("InitConnection succes.... ");

	RegistCmd(); //注册指令集
	logger.Info("RegistCmd succes.... ");

	nResult := StartListen()
	if nResult != nil {        //启动服务错误
		logger.Error("StartListen is failure =\n", nResult);
		os.Exit(1)
	}
	logger.Info("StartListen succes.... ", *ipPort);

	//主线程更新
	go mainFrameUpdate();

	//控制台指令
	mainConsole();


	mainClose();
}

func mainConsole() {
	reader := bufio.NewReader(os.Stdin)
	for WorldRunning {
		data, _, _ := reader.ReadLine()
		command := string(data)
		//log.Println("command", command,",a1= ",a1,",a2= ",a2)
		if command == "stop" {
			WorldRunning = false
			break;
		}
		time.Sleep(time.Second)
	}
}

//主线程更新
func mainFrameUpdate() {
	var (
		startTime int = 0
		endTime int   = 0
		timeout int   = 0
	)

	for {
		startTime = time.Now().Nanosecond()
		//更新涵数


		//超时记录
		endTime = time.Now().Nanosecond()
		timeout = endTime - startTime
		if (timeout > MainFrameUpdate*2) {
			logger.Info("mainFrameUpdate is Timeout[", MainFrameUpdate, "], from [", startTime, "] to [", endTime, "],timeout=", timeout)
		}
		time.Sleep((MainFrameUpdate))
		//有退出标记，退出主线程
		if !WorldRunning {
			break
		}
	}
	logger.Error("Start world server stop ......\n")
}

//正常退出
func mainClose() {
	if !WorldRunning {
		CloseConnection();
		//DbClose(DbConn);
	}
}

//加载配置
func initConfig() {

	logger.SetConsole(false)
	logger.SetRollingFile("log", logFileName, maxFileCount, maxFileSize, logger.KB)
	logger.SetLevel(logLevel)

	common.LoadConfig(ConfigFile)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			logger.Error(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	//	"sync/atomic"
	"syscall"
)

//----------------------------------------------- handle unix signals
func SignalProc() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM)

	for {
		msg := <-ch
		switch msg {
		case syscall.SIGHUP: // reload config        kill -1
			log.Println("[SIGHUP] reload config")

		case syscall.SIGTERM: // server close      kill -15
			//atomic.StoreInt32(&SIGTERM, 1)
			log.Println("[SIGTERM]server closed is start.....")
			ServerTerminate()
			log.Println("[SIGTERM]server closed is sucees.")
			os.Exit(0)
		}
	}
}

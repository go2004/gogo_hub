package helper

import (
	"log"
	"runtime"
)

func PrintPanicStack() {
	if x := recover(); x != nil {
		//	log.Printf(x)
		for i := 0; i < 10; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if ok {
				log.Printf("frame ", i, ":[func:", runtime.FuncForPC(funcName).Name(), ",file:", file, ",line:%v]\n", line)
			}
		}
	}
}

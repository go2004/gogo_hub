package helper

import (
	"runtime"
	//"log"
)

//---------------------------------------------------------- system routine
func SysRoutine() {
	runtime.GC()
//	log.Printf("== PERFORMANCE LOG ==")
//	log.Printf("Goroutine Count:", runtime.NumGoroutine())
//	log.Printf("GC Summary:", GCSummary())
}

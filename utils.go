package myrxgo

import (
	"crypto/rand"
	"fmt"
	"log"
	"runtime"
)

var IsDebug = true
var SystemPanicHandler func(rec interface{}, stack []byte) error

func init() {
	SystemPanicHandler = func(rec interface{}, stack []byte) error {
		err := fmt.Errorf("PANIC: %s\n%s", rec, stack)
		Errorf("PANIC: %s\n%s", rec, stack)
		return err
	}
}

func Debugf(fmt string, i ...interface{}) {
	if IsDebug {
		log.Printf(fmt, i...)
	}
}

func Errorf(fmt string, i ...interface{}) {
	log.Printf("ERROR "+fmt, i...)
}

func Try(f func(), except func(rec interface{}, stack []byte)) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			except(r, stack)
		}
	}()
	f()
}

func safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]

			Errorf("PANIC: %s\n%s", r, stack)
		}
	}()
	fn()
}

func safeGo(fn func(...interface{}), args ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				Errorf("PANIC: %s\n%s", err, stack)
			}
		}()

		fn(args...)
	}()
}

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

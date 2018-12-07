package myrxgo

import (
	"crypto/rand"
	"fmt"
	"log"
	"runtime"
)

func safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
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

				f := "PANIC: %s\n%s"
				log.Printf(f, err, stack)
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

package logger

import (
	"log"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func recovery() {
	if r := recover(); r != nil {
		log.Println(r)
	}
}

func Test_init(t *testing.T) {
	Debug("not panic")
	Init("release")
	Debug("not print")
}

func TestLogger(t *testing.T) {
	Init("debug")

	defer Sync()

	Debug("Debug")
	Debugf("Debugf: %v", "Debugging!")

	Info("Info")
	Infof("Infof: %v", "Hello World!")

	Warn("Warn")
	Warnf("Warnf: %v", "Warning!")

	Error("Error")
	Errorf("Errorf: %v", "Something went wrong!")

	testingFuncs := []func(){
		func() {
			DPanic("DPanic")
		},
		func() {
			DPanicf("DPanicf: %v", "Panic!")
		},
		func() {
			Panic("Panic")
		},
		func() {
			Panicf("Panicf: %v", "Panic Error!")
		},
		func() {
			// Fatal("Fatal")
		},
		func() {
			// Fatalf("Fatalf: %v", "Fatal Error!")
		},
	}

	wg.Add(len(testingFuncs))
	for _, f := range testingFuncs {
		go func(f func()) {
			defer recovery()
			defer wg.Done()
			f()
		}(f)
	}

	wg.Wait()
}

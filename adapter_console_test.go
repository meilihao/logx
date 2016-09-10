package logx

import (
	"fmt"
	"testing"
)

func testConsoleCalls(l *Logger) {
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
	//l.Panic("panic")
	//l.Fatal("fatal")
}

func TestConsole(t *testing.T) {
	log := NewLogger()
	log.SetFuncCallDepth(4)
	log.AddLogger("console", "")
	fmt.Println(log.GetLevel())
	testConsoleCalls(log)
}

func TestConsoleWithShortfile(t *testing.T) {
	log := NewLogger()
	log.SetFuncCallDepth(1)
	log.SetShortfile(true)
	log.AddLogger("console", "")
	fmt.Println(log.GetLevel())
	testConsoleCalls(log)
}

func TestConsoleWithLevel(t *testing.T) {
	log := NewLogger()
	log.SetLevel(LevelError)
	log.AddLogger("console", `{}`)
	testConsoleCalls(log)
}

func TestConsoleNoColor(t *testing.T) {
	log := NewLogger()
	log.AddLogger("console", `{"color":false}`)
	testConsoleCalls(log)
}

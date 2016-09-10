package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Logger struct {
	lock          sync.Mutex
	level         int
	isShortfile   bool
	funcCallDepth int
	msgChanLen    int64
	msgChan       chan *logMsg
	signalChan    chan string
	wg            sync.WaitGroup
	outputs       []*nameLogger
}

type nameLogger struct {
	Storer
	name string
}

func NewLogger() *Logger {
	l := new(Logger)

	l.level = LevelDebug
	l.funcCallDepth = 2
	l.signalChan = make(chan string, 1)

	return l
}

func (l *Logger) AddLogger(adapterName string, config ...string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	cfg := append(config, "{}")[0]
	if cfg == "" {
		cfg = "{}"
	}

	var storer Storer
	switch adapterName {
	case AdapterConsole:
		storer = newAdapterConsole()
	case AdapterFile:
		storer = newAdapterFile()
	case AdapterMultifile:
		storer = newAdapterMultifile()
	}

	if storer == nil {
		return fmt.Errorf("logx: unknown adaptername %q", adapterName)
	}

	err := storer.Init(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			fmt.Sprintf("logx: init adaptername(%s) error:%v", adapterName, err.Error()))
		return err
	}
	l.outputs = append(l.outputs, &nameLogger{name: adapterName, Storer: storer})
	return nil
}

func (l *Logger) DeleteLogger(adapterName string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	outputs := []*nameLogger{}
	for _, v := range l.outputs {
		if v.name == adapterName {
			v.Destroy()
		} else {
			outputs = append(outputs, v)
		}
	}
	if len(outputs) == len(l.outputs) {
		return fmt.Errorf("logx: unknown adaptername %q", adapterName)
	}

	l.outputs = outputs
	return nil
}

func (l *Logger) Debug(v ...interface{}) {
	if LevelDebug < l.level {
		return
	}

	l.writeMsg(LevelDebug, generateFmtStr(len(v)), v...)
}

func (l *Logger) Info(v ...interface{}) {
	if LevelInfo < l.level {
		return
	}

	l.writeMsg(LevelInfo, generateFmtStr(len(v)), v...)
}

func (l *Logger) Warn(v ...interface{}) {
	if LevelWarn < l.level {
		return
	}

	l.writeMsg(LevelWarn, generateFmtStr(len(v)), v...)
}

func (l *Logger) Error(v ...interface{}) {
	if LevelError < l.level {
		return
	}

	l.writeMsg(LevelError, generateFmtStr(len(v)), v...)
}

func (l *Logger) Panic(v ...interface{}) {
	if LevelPanic < l.level {
		return
	}

	l.writeMsg(LevelPanic, generateFmtStr(len(v)), v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	if LevelFatal < l.level {
		return
	}

	l.writeMsg(LevelFatal, generateFmtStr(len(v)), v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if LevelDebug < l.level {
		return
	}

	l.writeMsg(LevelDebug, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if LevelInfo < l.level {
		return
	}

	l.writeMsg(LevelInfo, format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if LevelWarn < l.level {
		return
	}

	l.writeMsg(LevelWarn, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if LevelError < l.level {
		return
	}

	l.writeMsg(LevelError, format, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	if LevelPanic < l.level {
		return
	}

	l.writeMsg(LevelPanic, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if LevelFatal < l.level {
		return
	}

	l.writeMsg(LevelFatal, format, v...)
}

func (l *Logger) ErrDebug(err error) {
	if err == nil || LevelDebug < l.level {
		return
	}

	l.writeMsg(LevelDebug, "%v", err)
}

func (l *Logger) ErrInfo(err error) {
	if err == nil || LevelInfo < l.level {
		return
	}

	l.writeMsg(LevelInfo, "%v", err)
}

func (l *Logger) ErrWarn(err error) {
	if err == nil || LevelWarn < l.level {
		return
	}

	l.writeMsg(LevelWarn, "%v", err)
}

func (l *Logger) ErrError(err error) {
	if err == nil || LevelError < l.level {
		return
	}

	l.writeMsg(LevelError, "%v", err)
}

func (l *Logger) ErrPanic(err error) {
	if err == nil || LevelPanic < l.level {
		return
	}

	l.writeMsg(LevelPanic, "%v", err)
}

func (l *Logger) ErrFatal(err error) {
	if err == nil || LevelFatal < l.level {
		return
	}

	l.writeMsg(LevelFatal, "%v", err)
}

func (l *Logger) writeMsg(level int, msg string, v ...interface{}) error {
	if len(v) == 0 {
		panic("logx: Empty Output")
		return nil
	}

	msg = fmt.Sprintf(msg, v...)
	when := time.Now()

	if l.funcCallDepth > 0 {
		_, file, line, ok := runtime.Caller(l.funcCallDepth)
		if !ok {
			file = "???"
			line = 0
		}

		if l.isShortfile {
			file = filepath.Base(file)
		}

		msg = "[" + file + ":" + strconv.FormatInt(int64(line), 10) + "] " + msg
	}

	msg = levelPrefix[level] + msg

	if l.msgChanLen > 0 {
		lm := logMsgPool.Get().(*logMsg)
		lm.level = level
		lm.msg = msg
		lm.when = when
		l.msgChan <- lm
	} else {
		l.writeToLoggers(when, msg, level)
	}

	switch level {
	case LevelPanic:
		panic(msg)
	case LevelFatal:
		os.Exit(1)
	}
	return nil
}

func (l *Logger) writeToLoggers(when time.Time, msg string, level int) {
	for _, v := range l.outputs {
		err := v.WriteMsg(when, msg, level)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"logx: write to adapter(%s) error:%v\n", v.name, err)
		}
	}
}

func (l *Logger) flush() {
	if l.msgChanLen > 0 {
		for {
			if len(l.msgChan) > 0 {
				lm := <-l.msgChan
				l.writeToLoggers(lm.when, lm.msg, lm.level)
				logMsgPool.Put(lm)
				continue
			}
			break
		}
	}
	for _, l := range l.outputs {
		l.Flush()
	}
}

func (l *Logger) Flush() {
	if l.msgChanLen > 0 {
		l.signalChan <- "flush"
		l.wg.Wait()
		l.wg.Add(1)
	} else {
		l.flush()
	}
}

func (l *Logger) Close() {
	if l.msgChanLen > 0 {
		l.signalChan <- "close"
		l.wg.Wait()
		close(l.msgChan)
	} else {
		l.flush()
		for _, v := range l.outputs {
			v.Destroy()
		}

		l.outputs = nil
	}
	close(l.signalChan)
}

func (l *Logger) Reset() {
	l.flush()

	for _, v := range l.outputs {
		v.Destroy()
	}

	l.outputs = nil
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) GetLevel() int {
	return l.level
}

func (l *Logger) SetFuncCallDepth(depth int) {
	l.funcCallDepth = depth
}

func (l *Logger) GetFuncCallDepth() int {
	return l.funcCallDepth
}

// default is false
func (l *Logger) SetShortfile(b bool) {
	l.isShortfile = b
}

func (l *Logger) GetShortfile() bool {
	return l.isShortfile
}

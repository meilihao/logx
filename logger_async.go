package logx

import (
	"sync"
	"time"
)

const (
	defaultAsyncMsgLen = 1e3
)

type logMsg struct {
	level int
	msg   string
	when  time.Time
}

var logMsgPool = &sync.Pool{
	New: func() interface{} {
		return &logMsg{}
	},
}

func (l *Logger) Async(length ...int64) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.msgChanLen > 0 {
		return
	}

	l.msgChanLen = append(length, defaultAsyncMsgLen)[0]

	if l.msgChanLen <= 0 {
		l.msgChanLen = defaultAsyncMsgLen
	}

	l.msgChan = make(chan *logMsg, l.msgChanLen)
	l.wg.Add(1)

	go l.startLogger()
}

func (l *Logger) startLogger() {
	gameOver := false

	for {
		select {
		case lm := <-l.msgChan:
			l.writeToLoggers(lm.when, lm.msg, lm.level)
			logMsgPool.Put(lm)
		case sg := <-l.signalChan:
			// Now should only send "flush" or "close" to l.signalChan
			l.flush()

			if sg == "close" {
				for _, l := range l.outputs {
					l.Destroy()
				}

				l.outputs = nil
				gameOver = true
			}

			l.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

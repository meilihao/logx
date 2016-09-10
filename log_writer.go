package logx

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type logWriter struct {
	sync.Mutex
	writer io.Writer
}

func newLogWriter(w io.Writer) *logWriter {
	return &logWriter{writer: w}
}

func (lw *logWriter) println(when time.Time, msg string) {
	lw.Lock()

	buf := msgBufPool.Get()
	b := buf.(*bytes.Buffer)
	b.Reset()

	b.WriteString(when.Format(timeLayout))
	b.WriteString(" ")
	b.WriteString(msg)
	b.WriteString("\n")
	lw.writer.Write(b.Bytes())

	msgBufPool.Put(b)

	lw.Unlock()
}

var msgBufPool = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

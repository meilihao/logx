package logx

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	AdapterMultifile = "multifile"
)

// A filesLogWriter manages several fileWriter
// filesLogWriter will write logs to the file in json configuration  and write the same level log to correspond file
// means if the file name in configuration is project.log filesLogWriter will create project.error.log/project.debug.log
// and write the error-level logs to project.error.log and write the debug-level logs to project.debug.log
// the rotate attribute also  acts like fileWriter
type multifileWriter struct {
	writers    []*fileWriter
	fullWriter *fileWriter
	Separate   []string `json:"separate"`
	IsFull     bool     `json:"full"`
}

var levelIndex map[int]int

// Init file logger with json config.
// jsonConfig like:
//	{
//	"filename":"logs/app.log",
//	"maxline":0,
//	"maxsize":0,
//	"daily":true,
//	"maxday":30,
//  "perm":0600,
//	"full":false,
//	"separate":["debug","info","warn","error","panic","fatal"],
//	}

func (w *multifileWriter) Init(jsonConfig string) error {
	err := json.Unmarshal([]byte(jsonConfig), w)
	if err != nil {
		return err
	}

	levelIndex = map[int]int{}
	for i, v := range w.Separate {
		level := GetLevelByName(v)
		if level == -1 {
			panic(fmt.Sprintf("unknown Level(%s)", v))
		}

		_, ok := levelIndex[level]
		if ok {
			panic(fmt.Sprintf("double Level(%s)", v))
		}

		levelIndex[level] = i
	}

	if len(levelIndex) == 0 {
		panic(fmt.Sprint("empty Separate Level"))
	}

	writer := newAdapterFile().(*fileWriter)
	err = writer.Init(jsonConfig)
	if err != nil {
		return err
	}
	w.fullWriter = writer

	jsonMap := map[string]interface{}{}
	json.Unmarshal([]byte(jsonConfig), &jsonMap)

	for _, v := range w.Separate {
		jsonMap["filename"] = w.fullWriter.filePrefix + "." + v + w.fullWriter.fileExt
		bs, _ := json.Marshal(jsonMap)
		writer := newAdapterFile().(*fileWriter)
		writer.Init(string(bs))
		w.writers = append(w.writers, writer)
	}

	if !w.IsFull {
		w.fullWriter = nil
	}

	return nil
}

func (w *multifileWriter) Destroy() {
	for i := 0; i < len(w.writers); i++ {
		if w.writers[i] != nil {
			w.writers[i].Destroy()
		}
	}
	if w.IsFull {
		w.fullWriter.Destroy()
	}
}

func (w *multifileWriter) WriteMsg(when time.Time, msg string, level int) error {
	if w.IsFull {
		w.fullWriter.WriteMsg(when, msg, level)
	}

	v, ok := levelIndex[level]
	if ok {
		w.writers[v].WriteMsg(when, msg, level)
	}

	return nil
}

func (w *multifileWriter) Flush() {
	for i := 0; i < len(w.writers); i++ {
		if w.writers[i] != nil {
			w.writers[i].Flush()
		}
	}
	if w.IsFull {
		w.fullWriter.Flush()
	}
}

func newAdapterMultifile() Storer {
	return &multifileWriter{IsFull: true}
}

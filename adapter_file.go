package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	AdapterFile = "file"
)

// fileWriter implements LoggerInterface.
// It writes messages by lines limit, file size limit, or time frequency.
type fileWriter struct {
	sync.RWMutex // write log order by order and  atomic incr maxLineCurLine and maxSizeCurSize
	// The opened file
	Filename string `json:"filename"`
	file     *os.File

	// Rotate at line
	MaxLine        int `json:"maxline"`
	maxLineCurLine int

	// Rotate at size
	MaxSize        int `json:"maxsize"`
	maxSizeCurSize int

	// Rotate daily
	Daily         bool  `json:"daily"`
	MaxDay        int64 `json:"maxday"`
	dailyOpenDate int
	dailyOpenTime time.Time

	Perm string `json:"perm"`

	filePrefix, fileExt string // like "project.log", project is filePrefix and .log is fileExt
}

// newAdapterFile create a FileWriter returning as LoggerInterface.
func newAdapterFile() Storer {
	w := &fileWriter{
		Filename: "app.log",
		Daily:    true,
		MaxDay:   30,
		Perm:     "0644",
	}
	return w
}

// Init file logger with json config.
// jsonConfig like:
//	{
//	"filename":"logs/app.log",
//	"maxline":10000,
//	"maxsize":1024,
//	"daily":true,
//	"maxday":15,
//  "perm":"0600"
//	}
func (w *fileWriter) Init(jsonConfig string) error {
	err := json.Unmarshal([]byte(jsonConfig), w)
	if err != nil {
		return err
	}

	w.fileExt = filepath.Ext(w.Filename)
	w.filePrefix = strings.TrimSuffix(w.Filename, w.fileExt)
	if w.fileExt == "" {
		w.fileExt = ".log"
	}

	err = w.startLogger()
	return err
}

// start file logger. create log file and set to locker-inside file writer.
func (w *fileWriter) startLogger() error {
	file, err := w.createLogFile()
	if err != nil {
		return err
	}
	if w.file != nil {
		w.file.Close()
	}
	w.file = file
	return w.initFd()
}

func (w *fileWriter) needRotate(day int) bool {
	return (w.MaxLine > 0 && w.maxLineCurLine >= w.MaxLine) ||
		(w.MaxSize > 0 && w.maxSizeCurSize >= w.MaxSize) ||
		(w.Daily && day != w.dailyOpenDate)

}

// WriteMsg write logger message into file.
func (w *fileWriter) WriteMsg(when time.Time, msg string, level int) error {
	msg = when.Format(timeLayout) + " " + msg + "\n"
	if w.Daily {
		w.RLock()
		if w.needRotate(when.Day()) {
			w.RUnlock()
			w.Lock()

			if err := w.doRotate(when); err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", w.Filename, err)
			}

			w.Unlock()
		} else {
			w.RUnlock()
		}
	}

	w.Lock()
	_, err := w.file.Write([]byte(msg))
	if err == nil {
		w.maxLineCurLine++
		w.maxSizeCurSize += len(msg)
	}
	w.Unlock()
	return err
}

func (w *fileWriter) createLogFile() (*os.File, error) {
	// Open the log file
	perm, err := strconv.ParseUint(w.Perm, 8, 32)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(w.Filename,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(w.Filename, os.FileMode(perm))
	}
	return fd, err
}

func (w *fileWriter) initFd() error {
	fd := w.file
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s\n", err)
	}
	w.maxSizeCurSize = int(fInfo.Size())
	w.dailyOpenTime = time.Now()
	w.dailyOpenDate = w.dailyOpenTime.Day()
	w.maxLineCurLine = 0
	if w.Daily {
		go w.dailyRotate(w.dailyOpenTime)
	}
	if w.maxSizeCurSize > 0 {
		count, err := w.lines()
		if err != nil {
			return err
		}
		w.maxLineCurLine = count
	}
	return nil
}

// rotate at 00:00
func (w *fileWriter) dailyRotate(openTime time.Time) {
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	select {
	case <-tm.C:
		w.Lock()
		now := time.Now()
		if w.needRotate(now.Day()) {
			if err := w.doRotate(now); err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", w.Filename, err)
			}
		}
		w.Unlock()
	}
	tm.Stop()
}

func (w *fileWriter) lines() (int, error) {
	fd, err := os.Open(w.Filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32*1<<10) // 32k
	count := 0

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], []byte{'\n'})

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// DoRotate means it need to write file in new file.
// new file name like xx.2016-01-02.log (daily) or xx.2016-01-02.001.log (by line or size)
func (w *fileWriter) doRotate(when time.Time) error {
	newFilename := ""

	// file exists
	_, err := os.Lstat(w.Filename)
	if err != nil {
		//even if the file is not exist or other ,we should RESTART the logger
		goto RESTART_LOGGER
	}

	newFilename = w.getNewFilname(when)
	// return error if the last file checked still existed
	if newFilename == "" {
		return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.Filename)
	}

	// close fileWriter before rename
	w.file.Close()

	// Rename the file to its new found name
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(w.Filename, newFilename)

	if err != nil {
		return fmt.Errorf("Rotate: %s\n", err)
	}

RESTART_LOGGER:
	startLoggerErr := w.startLogger()
	if startLoggerErr != nil {
		return fmt.Errorf("Rotate StartLogger: %s\n", startLoggerErr)
	}

	go w.deleteOldLog()

	return nil

}

// Find the next available number
func (w *fileWriter) getNewFilname(when time.Time) string {
	num := 1
	fName := ""
	var err error

	if w.MaxLine > 0 || w.MaxSize > 0 {
		for ; err == nil; num++ {
			fName = w.filePrefix + fmt.Sprintf(".%s.%03d%s", when.Format("2006-01-02"), num, w.fileExt)
			_, err = os.Lstat(fName)
		}
	} else {
		fName = w.filePrefix + fmt.Sprintf(".%s%s", w.dailyOpenTime.Format("2006-01-02"), w.fileExt)
		_, err = os.Lstat(fName)
		for ; err == nil; num++ {
			fName = w.filePrefix + fmt.Sprintf(".%s.%03d%s", w.dailyOpenTime.Format("2006-01-02"), num, w.fileExt)
			_, err = os.Lstat(fName)
		}
	}

	return fName
}

func (w *fileWriter) deleteOldLog() {
	if w.MaxDay <= 0 {
		return
	}

	now := time.Now()
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}

		if !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(w.MaxDay)).Before(now) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.filePrefix)) &&
				strings.HasSuffix(filepath.Base(path), w.fileExt) {
				os.Remove(path)
			}
		}
		return
	})
}

// Destroy close the file description, close file writer.
func (w *fileWriter) Destroy() {
	w.Lock()
	w.file.Close()
	w.Unlock()
}

// Flush flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (w *fileWriter) Flush() {
	w.Lock()
	w.file.Sync()
	w.Unlock()
}

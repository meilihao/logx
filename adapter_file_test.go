package logx

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFilePerm(t *testing.T) {
	log := NewLogger()
	// use 0666 as test perm cause the default umask is 022
	log.AddLogger("file", `{"filename":"test.log", "perm": "0666"}`)
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	file, err := os.Stat("test.log")
	if err != nil {
		t.Fatal(err)
	}
	if file.Mode() != 0666 {
		t.Fatal("unexpected log file permission")
	}
	os.Remove("test.log")
}

func TestFile1(t *testing.T) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test.log"}`)
	log.Debug("debug")
	f, err := os.Open("test.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = 1
	if lineNum != expected {
		t.Fatal(lineNum, "not "+strconv.Itoa(expected)+" lines")
	}
	os.Remove("test.log")
}

func TestFileRotate_01(t *testing.T) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test3.log","maxline":4}`)
	for i := 0; i <= 5; i++ {
		log.Debug("debug")
	}
	rotateName := "test3" + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), 1) + ".log"
	b, err := exists(rotateName)
	if !b || err != nil {
		os.Remove("test3.log")
		t.Fatal("rotate not generated")
	}
	os.Remove(rotateName)
	os.Remove("test3.log")
}

func TestFileRotate_02(t *testing.T) {
	fn1 := "rotate_day.log"
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	testFileRotate(t, fn1, fn2)
}

func TestFileRotate_03(t *testing.T) {
	fn1 := "rotate_day.log"
	fn := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	os.Create(fn)
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileRotate(t, fn1, fn2)
	os.Remove(fn)
}

func TestFileRotate_04(t *testing.T) {
	fn1 := "rotate_day.log"
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	testFileDailyRotate(t, fn1, fn2)
}

func TestFileRotate_05(t *testing.T) {
	fn1 := "rotate_day.log"
	fn := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".log"
	os.Create(fn)
	fn2 := "rotate_day." + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".001.log"
	testFileDailyRotate(t, fn1, fn2)
	os.Remove(fn)
}

// do one rotate
func testFileRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileWriter{
		Daily:  true,
		MaxDay: 7,
		Perm:   "0660",
	}

	fw.Init(fmt.Sprintf(`{"filename":"%v","maxday":1}`, fn1))
	fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
	fw.dailyOpenDate = fw.dailyOpenTime.Day()

	fw.WriteMsg(time.Now(), "this is a msg for test", LevelDebug)

	fw.Destroy()
	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.FailNow()
		}
		os.Remove(file)
	}
}

func testFileDailyRotate(t *testing.T, fn1, fn2 string) {
	fw := &fileWriter{
		Daily:  true,
		MaxDay: 7,
		Perm:   "0660",
	}

	fw.Init(fmt.Sprintf(`{"filename":"%v","maxday":1}`, fn1))
	fw.dailyOpenTime = time.Now().Add(-24 * time.Hour)
	fw.dailyOpenDate = fw.dailyOpenTime.Day()

	today, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), fw.dailyOpenTime.Location())
	today = today.Add(-1 * time.Second)
	fw.dailyRotate(today)
	time.Sleep(2 * time.Second)

	fw.Destroy()

	for _, file := range []string{fn1, fn2} {
		_, err := os.Stat(file)
		if err != nil {
			t.FailNow()
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			t.FailNow()
		}
		if len(content) > 0 {
			t.FailNow()
		}
		os.Remove(file)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func BenchmarkFile(b *testing.B) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileWithoutCallDepth(b *testing.B) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test4.log"}`)
	log.SetFuncCallDepth(0)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsynchronous(b *testing.B) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test4.log"}`)
	log.Async(1000)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileAsynchronousWithoutCallDepth(b *testing.B) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test4.log"}`)
	log.SetFuncCallDepth(0)
	log.Async(100)
	for i := 0; i < b.N; i++ {
		log.Debug("debug")
	}
	os.Remove("test4.log")
}

func BenchmarkFileOnGoroutine(b *testing.B) {
	log := NewLogger()
	log.AddLogger("file", `{"filename":"test4.log"}`)
	for i := 0; i < b.N; i++ {
		go log.Debug("debug")
	}
	os.Remove("test4.log")
}

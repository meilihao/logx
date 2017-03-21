package logx

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMutifile_NoFull(t *testing.T) {
	log := NewLogger()
	log.AddLogger("multifile", `{"filename":"test.log","separate":["debug", "info"]}`)
	log.Debug("debug")
	log.Info("info")

	fns := []string{"debug", "info"}
	name := "test"
	suffix := ".log"
	for _, fn := range fns {
		file := name + "." + fn + suffix
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		b := bufio.NewReader(f)
		lineNum := 0
		lastLine := ""
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				lastLine = string(line)
				lineNum++
			}
		}
		var expected = 1
		if lineNum != expected {
			t.Fatal(file, "has", lineNum, "lines not "+strconv.Itoa(expected)+" lines")
		}
		if lineNum == 1 {
			if !strings.Contains(lastLine, fn) {
				t.Fatal(file + " " + lastLine + " not contains the log msg " + fn)
			}
		}
		os.Remove(file)
	}
}

func TestMutifile_Full(t *testing.T) {
	log := NewLogger()
	log.AddLogger("multifile", `{"filename":"test.log","full":true,"separate":["debug", "info"]}`)
	log.Debug("debug")
	log.Info("info")

	fns := []string{"", "debug", "info"}
	expected := []int{2, 1, 1}

	name := "test"
	suffix := ".log"
	for i, fn := range fns {
		var file string
		if fn == "" {
			file = name + suffix
		} else {
			file = name + "." + fn + suffix
		}

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		b := bufio.NewReader(f)
		lineNum := 0
		lastLine := ""
		for {
			line, _, err := b.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				lastLine = string(line)
				lineNum++
			}
		}

		if lineNum != expected[i] {
			t.Fatal(file, "has", lineNum, "lines not "+strconv.Itoa(expected[i])+" lines")
		}
		if fn != "" {
			if !strings.Contains(lastLine, fn) {
				t.Fatal(file + " " + lastLine + " not contains the log msg " + fn)
			}
		}
		os.Remove(file)
	}
}

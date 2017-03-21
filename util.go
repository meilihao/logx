package logx

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	timeLayout = "2006-01-02 15:04:05"
)

func SetTimeLayout(l string) {
	timeLayout = l
}

func generateFmtStr(n int) string {
	return strings.TrimRight(strings.Repeat("%v ", n), " ")
}

var LevelMap = map[string]int{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
	"panic": LevelPanic,
	"fatal": LevelFatal,
}

func GetLevelByName(name string) int {
	v, ok := LevelMap[name]
	if ok {
		return v
	} else {
		panic(fmt.Sprintf("unknown Level(%s)", v))
	}
}

func splitFilename(s string) (fileName, fileExt string) {
	fileExt = filepath.Ext(s)
	fileName = strings.TrimSuffix(s, fileExt)
	if fileExt == "" {
		fileExt = ".log"
	}

	return
}

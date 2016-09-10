package logx

import "strings"

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
		return -1
	}
}

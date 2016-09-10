// Copyright 2016 meilihao. All Rights Reserved.

package logx

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

var levelPrefix = [LevelFatal + 1]string{"[D] ", "[I] ", "[W] ", "[E] ", "[P] ", "[F] "}
var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
	defaultLogger.SetFuncCallDepth(3)
	defaultLogger.AddLogger(AdapterConsole)
}

func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

func SetOutput(l *Logger) {
	if l == nil {
		panic("logx: invalid Logger")
	}
	defaultLogger = l
}

// --- output
func Debug(v ...interface{}) {
	defaultLogger.Debugf(generateFmtStr(len(v)), v...)
}

func Info(v ...interface{}) {
	defaultLogger.Infof(generateFmtStr(len(v)), v...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warnf(generateFmtStr(len(v)), v...)
}

func Error(v ...interface{}) {
	defaultLogger.Errorf(generateFmtStr(len(v)), v...)
}

func Panic(v ...interface{}) {
	defaultLogger.Panicf(generateFmtStr(len(v)), v...)
}

func Fatal(v ...interface{}) {
	defaultLogger.Fatalf(generateFmtStr(len(v)), v...)
}

func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	defaultLogger.Panicf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

func ErrDebug(err error) {
	if err == nil {
		return
	}

	defaultLogger.Debugf("%v", err)
}

func ErrInfo(err error) {
	if err == nil {
		return
	}

	defaultLogger.Infof("%v", err)
}

func ErrWarn(err error) {
	if err == nil {
		return
	}

	defaultLogger.Warnf("%v", err)
}

func ErrError(err error) {
	if err == nil {
		return
	}

	defaultLogger.Errorf("%v", err)
}

func ErrPanic(err error) {
	if err == nil {
		return
	}

	defaultLogger.Panicf("%v", err)
}

func ErrFatal(err error) {
	if err == nil {
		return
	}

	defaultLogger.Fatalf("%v", err)
}

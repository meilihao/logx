package logx

import (
	"encoding/json"
	"os"
	"runtime"
	"time"
)

const (
	AdapterConsole = "console"
)

type consoleWriter struct {
	lg    *logWriter
	Color bool `json:"color"`
}

func newAdapterConsole() Storer {
	w := &consoleWriter{
		lg:    newLogWriter(os.Stdout),
		Color: runtime.GOOS != "windows",
	}

	return w
}

func (c *consoleWriter) Init(jsonConfig string) error {
	err := json.Unmarshal([]byte(jsonConfig), c)
	if runtime.GOOS == "windows" {
		c.Color = false
	}
	return err
}

// WriteMsg write message in console.
func (c *consoleWriter) WriteMsg(when time.Time, msg string, level int) error {
	if c.Color {
		msg = colors[level](msg)
	}
	c.lg.println(when, msg)
	return nil
}

func (c *consoleWriter) Destroy() {

}

func (c *consoleWriter) Flush() {

}

// --- color
type brush func(string) string

// newBrush return a fix color Brush
func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;37"), // Debug            white
	newBrush("1;34"), // Info             blue
	newBrush("1;33"), // Warn             yellow
	newBrush("1;31"), // Error            red
	newBrush("1;32"), // Panic            green
	newBrush("1;36"), // Fatal            cyan
}

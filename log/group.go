package log

// usage:
// log.NewGroupLogger(log.LevelInformational).Use(map[string]string {
// 	"console": "",
// 	"file": `{"file": "log.log"}`,
// })
//

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gitwillsky/slimgo/utils"
	"strings"
)

var std = NewGroupLogger(LevelInformational)

func SetLevel(level LogLevel) {
	std.SetLevel(level)
}

func GetLevel() LogLevel {
	return std.GetLevel()
}

func Alert(a ...interface{}) {
	msg := fmt.Sprintf(makeFormats(a), a...)
	std.Message(msg, LevelAlert)
	std.Close()
	os.Exit(1)
}

func Alertf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	std.Message(msg, LevelAlert)
	std.Close()
	os.Exit(1)
}

func Debug(a ...interface{}) {
	msg := fmt.Sprintf(makeFormats(a), a...)
	std.Message(msg, LevelDebug)
}

func Debugf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	std.Message(msg, LevelDebug)
}

func Error(a ...interface{}) {
	msg := fmt.Sprintf(makeFormats(a), a...)
	std.Message(msg, LevelError)
}

func Errorf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	std.Message(msg, LevelError)
}

func Info(a ...interface{}) {
	msg := fmt.Sprintf(makeFormats(a), a...)
	std.Message(msg, LevelInformational)
}

func Infof(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	std.Message(msg, LevelInformational)
}

func Warning(a ...interface{}) {
	msg := fmt.Sprintf(makeFormats(a), a...)
	std.Message(msg, LevelWarning)
}

func Warningf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	std.Message(msg, LevelWarning)
}

func SetOutput(loggers map[string]string) {
	std.Use(loggers)
}

func Close() {
	std.Close()
}

type groupLogger struct {
	implements  map[string]ILogger
	outputs     map[string]ILogger
	messages    chan *message
	level       LogLevel
	fileInfoLen int
	waitGroup   *sync.WaitGroup
}

func NewGroupLogger(level LogLevel) *groupLogger {
	g := &groupLogger{
		implements: implements,
		messages:   make(chan *message, 10),
		level:      level,
		waitGroup:  new(sync.WaitGroup),
	}

	// make sure console logger loaded
	Register("console", newConsole())

	// default logger
	g.Use(map[string]string{
		"console": "",
	})
	go g.start()
	return g
}

// Use use log implement to output log
// log name:config
func (g *groupLogger) Use(loggers map[string]string) {
	g.Close()
	g.outputs = make(map[string]ILogger, 4)
	for name, config := range loggers {
		impl, ok := g.implements[name]
		if !ok {
			fmt.Printf("log implement [%s] not found \n", name)
			continue
		}

		if err := impl.Init(config); err != nil {
			fmt.Printf("init log implement [%s] failed, %s", err.Error())
			continue
		}

		g.outputs[name] = impl
	}
}

func (g *groupLogger) SetLevel(level LogLevel) {
	g.level = level
}

func (g *groupLogger) GetLevel() LogLevel {
	level := g.level
	return level
}

type message struct {
	src      string
	dir      string
	filename string
	line     string
	level    LogLevel
}

// imp log interface method Message
func (g *groupLogger) Message(msg string, level LogLevel) error {
	if level > g.level {
		return nil
	}

	ms := new(message)
	ms.src = msg

	_, file, line, _ := runtime.Caller(2)
	ms.dir, ms.filename = path.Split(file)
	ms.line = strconv.Itoa(line)

	ms.level = level

	g.waitGroup.Add(1)
	g.messages <- ms

	return nil
}

// flush all chan data.
func (g *groupLogger) Flush() {
	for _, out := range g.outputs {
		out.Flush()
	}
}

func (g *groupLogger) format(msg *message) string {
	index := strings.LastIndex(msg.dir, "/src/")
	if index > 0 {
		msg.dir = msg.dir[index+5:]
	}
	fileInfo := msg.dir + msg.filename + ":" + msg.line
	// align right
	if len(fileInfo) < g.fileInfoLen {
		blank := make([]byte, g.fileInfoLen-len(fileInfo))
		for i := 0; i < g.fileInfoLen-len(fileInfo); i++ {
			blank[i] = ' '
		}
		fileInfo = utils.BytesToString(&blank) + fileInfo
	} else {
		g.fileInfoLen = len(fileInfo)
	}

	levelStr := ""
	switch msg.level {
	case LevelInformational:
		levelStr = "INFO "
	case LevelDebug:
		levelStr = "DEBUG"
	case LevelWarning:
		levelStr = "WARN "
	case LevelError:
		levelStr = "ERROR"
	case LevelAlert:
		levelStr = "ALERT"
		//case LevelNotice:
		//	levelStr = "NOTICE"
		//case LevelEmergency:
		//	levelStr = "Emergency"
		//case LevelCritical:
		//	levelStr = "CRITICAL"
	}

	return fmt.Sprintf("%s  %s [%s] : %s",
		time.Now().Format("2006-01-02 15:04:05"), levelStr, fileInfo, msg.src)
}

// close log
func (g *groupLogger) Close() {
	// wait all message write
	g.waitGroup.Wait()

	// all outputs flush and close.
	for _, out := range g.outputs {
		out.Flush()
		out.Close()
	}
}

func (g *groupLogger) start() {
	for {
		msg := <-g.messages
		for _, out := range g.outputs {
			if err := out.Message(g.format(msg), msg.level); err != nil {
				fmt.Println("ERROR, unable to WriteMsg (while closing log):", err)
			}
		}

		g.waitGroup.Done()
	}
}

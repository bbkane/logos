package logos_test

import (
	"go.bbkane.com/gocolor"
	"go.bbkane.com/logos"
	"go.uber.org/zap"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// https://blog.golang.org/examples
func Example() {
	// See https://github.com/natefinch/lumberjack for more options
	var lumberjackLogger *lumberjack.Logger = &lumberjack.Logger{
		Filename:   "/tmp/testlog.jsonl",
		MaxSize:    1, // megabytes
		MaxAge:     0,
		MaxBackups: 0,
		LocalTime:  true,
		Compress:   false,
	}
	color, err := gocolor.Prepare(true)
	if err != nil {
		panic(err)
	}
	l := logos.New(
		logos.NewBBKaneZapLogger(lumberjackLogger, zap.DebugLevel, "v1.0.0"),
		color,
	)
	defer l.Sync()
	l.LogOnPanic()
	l.Infow(
		"Now we're logging :)",
		"key", "value",
		"otherkey", "othervalue",
	)
}

package logos_test

import (
	"github.com/bbkane/logos"
	"go.uber.org/zap"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// https://blog.golang.org/examples
func Example() {
	// See https://github.com/natefinch/lumberjack for more options
	var lumberjackLogger *lumberjack.Logger = &lumberjack.Logger{
		Filename: "/tmp/testlog.jsonl",
		MaxSize:  1, // megabytes
	}
	sk := logos.NewLogger(
		logos.NewZapSugaredLogger(lumberjackLogger, zap.DebugLevel, "v1.0.0"),
	)
	defer sk.Sync()
	sk.LogOnPanic()
	sk.Infow(
		"Now we're logging :)",
		"key", "value",
		"otherkey", "othervalue",
	)
	// Output:
	// INFO: Now we're logging :)
	//   key: "value"
	//   otherkey: "othervalue"
}

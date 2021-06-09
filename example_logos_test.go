package logos_test

import (
	"os"

	"github.com/bbkane/logos"
	"go.uber.org/zap"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// https://blog.golang.org/examples
func Example() {
	// intialize a more useful lumberjack.Logger with:
	//   https://github.com/natefinch/lumberjack
	var lumberjackLogger *lumberjack.Logger = nil
	sk := logos.NewSugarKane(lumberjackLogger, os.Stderr, os.Stdout, zap.DebugLevel, "v1.0.0")
	defer sk.Sync()
	sk.LogOnPanic()
	sk.Infow(
		"Now we're logging :)",
		"key", "value",
	)
	// Output:
	// INFO: Now we're logging :)
	//   key: "value"
}

package logos

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// printw formats and prints a msg and keys and values to a stream.
// Useful when you need to show info but you don't have a log
func printw(fp *os.File, level string, msg string, keysAndValues ...interface{}) {
	msg = level + ": " + msg
	length := len(keysAndValues)
	if length%2 != 0 {
		panic(fmt.Sprintf("len() not even - keysAndValues: %#v\n", keysAndValues))
	}

	keys := make([]string, length/2)
	values := make([]interface{}, length/2)
	for i := 0; i < length/2; i++ {
		keys[i] = keysAndValues[i*2].(string)
		values[i] = keysAndValues[i*2+1]
	}

	fmtStr := msg + "\n"
	for i := 0; i < length/2; i++ {
		fmtStr += ("  " + keys[i] + ": %#v\n")
	}

	fmtStr += "\n"
	fmt.Fprintf(fp, fmtStr, values...)
}

// Logger is a very opinionated wrapper around a uber/zap sugared logger
// It's designed primarily to simultaneously print "pretty-enough" input for a
// user and useful enough info to a lumberjack logger
// It should really only be used with simple key/value pairs
// It's designed to be fairly easily swappable with the sugared logger
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// Infow prints a message and keys and values with INFO level
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
	printw(os.Stdout, "INFO", msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	printw(os.Stdout, "INFO", msg, keysAndValues...)
}

// Errorw prints a message and keys and values with INFO level
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
	printw(os.Stderr, "ERROR", msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	printw(os.Stderr, "ERROR", msg, keysAndValues...)
}

// Debugw prints keys and values only to the log, not to the user
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Sync syncs the underlying logger
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// LogOnPanic tries to log a panic. It should be called at the start of each
// goroutine. See panic and recover docs
func (l *Logger) LogOnPanic() {
	stackTraceSugar := l.logger.
		WithOptions(
			zap.AddStacktrace(zap.PanicLevel),
		).
		Sugar()
	if err := recover(); err != nil {
		stackTraceSugar.Panicw(
			"panic found!",
			"err", err,
		)
	}
}

// NewLogger builds a new Logger
func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{
		// logger is useful for syncs and panics
		logger: zapLogger,
		// sugar is a wrapper to call the things we actually care about :)
		sugar: zapLogger.WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}

// NewZapSugaredLogger builds a zap.SugaredLogger configured with settings I like
// If lumberjackLogger == nil, then it returns an No-op logger,
// which can be useful when you want to use the library, but not have a log file
func NewZapSugaredLogger(lumberjackLogger *lumberjack.Logger, lvl zapcore.LevelEnabler, appVersion string) *zap.Logger {
	if lumberjackLogger == nil {
		return zap.NewNop()
	}
	encoderConfig := zapcore.EncoderConfig{
		// prefix shared keys with '_' so they show up first when keys are alphabetical
		TimeKey:        "_timestamp",
		LevelKey:       "_level",
		NameKey:        "_name", // TODO: what is this?
		CallerKey:      "_caller",
		FunctionKey:    "_function", // zapcore.OmitKey,
		MessageKey:     "_msg",
		StacktraceKey:  "_stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	customCommonFields := zap.Fields(
		zap.Int("_pid", os.Getpid()),
		zap.String("_version", appVersion),
	)

	jsonCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(lumberjackLogger),
		lvl,
	)

	logger := zap.New(
		jsonCore,
		zap.AddCaller(),
		// Using errors package to get better stack traces
		// zap.AddStacktrace(stackTraceLvl),
		customCommonFields,
	)
	return logger
}

package logos

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// printw formats and prints a msg and keys and values to a stream.
// panics if keysAndValues doesn't have an even length
func printw(w io.Writer, level string, msg string, keysAndValues ...interface{}) {
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
	fmt.Fprintf(w, fmtStr, values...)
}

// Logger is a very opinionated wrapper around a uber/zap sugared logger
// It's designed primarily to simultaneously print "pretty-enough" input for a
// user and useful enough info to a lumberjack logger
// It should really only be used with simple key/value pairs
// It's designed to be fairly easily swappable with the sugared logger
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	stderr io.Writer
	stdout io.Writer
}

// Infow prints to stdout and the log
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
	printw(l.stdout, "INFO", msg, keysAndValues...)
}

// Infow prints to stdout
func Infow(msg string, keysAndValues ...interface{}) {
	printw(os.Stdout, "INFO", msg, keysAndValues...)
}

// Errorw prints to stderr and the log
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
	printw(l.stderr, "ERROR", msg, keysAndValues...)
}

// Errorw prints to stderr
func Errorw(msg string, keysAndValues ...interface{}) {
	printw(os.Stderr, "ERROR", msg, keysAndValues...)
}

// Debugw prints only to the log
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

// LoggerOpt allows customizations to New
type LoggerOpt func(*Logger)

// WithStderr overrides stderr for the logger
func WithStderr(stderr io.Writer) LoggerOpt {
	return func(l *Logger) {
		l.stderr = stderr
	}
}

// WithStdout overrides stdout for the logger
func WithStdout(stdout io.Writer) LoggerOpt {
	return func(l *Logger) {
		l.stdout = stdout
	}
}

// NewNop returns a no-op Logger. It never writes logs or prints
func NewNop() *Logger {
	return New(zap.NewNop(), WithStderr(io.Discard), WithStdout(io.Discard))
}

// New builds a new Logger
func New(logger *zap.Logger, opts ...LoggerOpt) *Logger {

	l := &Logger{
		logger: logger,
		sugar:  logger.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		stderr: nil,
		stdout: nil,
	}
	for _, opt := range opts {
		opt(l)
	}
	if l.stderr == nil {
		WithStderr(os.Stderr)(l)
	}
	if l.stdout == nil {
		WithStdout(os.Stdout)(l)
	}

	return l
}

// NewDeterministicZapLogger saves only levels, names, and messages for testing purposes
func NewDeterministicZapLogger(w io.Writer) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		// prefix shared keys with '_' so they show up first when keys are alphabetical
		TimeKey:          zapcore.OmitKey,
		LevelKey:         "_level",
		NameKey:          "_name", // TODO: what is this?
		CallerKey:        zapcore.OmitKey,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "_msg",
		StacktraceKey:    zapcore.OmitKey,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       nil,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     nil,
		EncodeName:       nil,
		ConsoleSeparator: "",
	}
	jsonCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(w),
		zap.DebugLevel,
	)
	logger := zap.New(
		jsonCore,
	)
	return logger
}

// NewBBKaneZapLogger builds a zap.SugaredLogger configured with settings I like
func NewBBKaneZapLogger(lumberjackLogger *lumberjack.Logger, lvl zapcore.LevelEnabler, appVersion string) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		// prefix shared keys with '_' so they show up first when keys are alphabetical
		TimeKey:          "_timestamp",
		LevelKey:         "_level",
		NameKey:          "_name", // TODO: what is this?
		CallerKey:        "_caller",
		FunctionKey:      "_function", // zapcore.OmitKey,
		MessageKey:       "_msg",
		StacktraceKey:    "_stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeName:       nil,
		ConsoleSeparator: "",
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

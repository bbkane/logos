package logos

import (
	"fmt"
	"io"
	"os"

	"go.bbkane.com/gocolor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// printw formats and prints a msg and keys and values to a stream.
// panics if keysAndValues doesn't have an even length
func printw(w io.Writer, color gocolor.Color, coloredLevel string, msg string, keysAndValues ...interface{}) {
	length := len(keysAndValues)
	if length%2 != 0 {
		panic(fmt.Sprintf("len() not even - keysAndValues: %#v\n", keysAndValues))
	}

	msg = coloredLevel + ": " + msg

	keys := make([]string, length/2)
	values := make([]interface{}, length/2)
	for i := 0; i < length/2; i++ {
		keys[i] = color.Add(color.Bold, keysAndValues[i*2].(string))
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
	color  gocolor.Color
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	stderr io.Writer
	stdout io.Writer
}

// Infow prints to stdout and the log
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
	coloredLevel := l.color.Add(l.color.Bold+l.color.FgGreenBright, "INFO")
	printw(l.stdout, l.color, coloredLevel, msg, keysAndValues...)
}

// Errorw prints to stderr and the log
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
	coloredLevel := l.color.Add(l.color.Bold+l.color.FgRedBright, "ERROR")
	printw(l.stderr, l.color, coloredLevel, msg, keysAndValues...)
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
	return New(zap.NewNop(), gocolor.NewEmpty(), WithStderr(io.Discard), WithStdout(io.Discard))
}

// New builds a new Logger. color should be initialized.
func New(logger *zap.Logger, color gocolor.Color, opts ...LoggerOpt) *Logger {

	l := &Logger{
		color:  color,
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
		TimeKey:             zapcore.OmitKey,
		LevelKey:            "_level",
		NameKey:             "_name", // TODO: what is this?
		CallerKey:           zapcore.OmitKey,
		FunctionKey:         zapcore.OmitKey,
		MessageKey:          "_msg",
		StacktraceKey:       zapcore.OmitKey,
		LineEnding:          zapcore.DefaultLineEnding,
		EncodeLevel:         zapcore.CapitalLevelEncoder,
		EncodeTime:          nil,
		EncodeDuration:      zapcore.StringDurationEncoder,
		EncodeCaller:        nil,
		EncodeName:          nil,
		ConsoleSeparator:    "",
		SkipLineEnding:      false,
		NewReflectedEncoder: nil,
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

// NewBBKaneZapLogger builds a zap.SugaredLogger configured with settings I like. As a special case,
// if lumberjackLogger == nil, then returns zap.newNop
func NewBBKaneZapLogger(lumberjackLogger *lumberjack.Logger, lvl zapcore.LevelEnabler, appVersion string) *zap.Logger {
	if lumberjackLogger == nil {
		return zap.NewNop()
	}
	encoderConfig := zapcore.EncoderConfig{
		// prefix shared keys with '_' so they show up first when keys are alphabetical
		TimeKey:             "_timestamp",
		LevelKey:            "_level",
		NameKey:             "_name", // TODO: what is this?
		CallerKey:           "_caller",
		FunctionKey:         "_function", // zapcore.OmitKey,
		MessageKey:          "_msg",
		StacktraceKey:       "_stacktrace",
		LineEnding:          zapcore.DefaultLineEnding,
		EncodeLevel:         zapcore.CapitalLevelEncoder,
		EncodeTime:          zapcore.ISO8601TimeEncoder,
		EncodeDuration:      zapcore.StringDurationEncoder,
		EncodeCaller:        zapcore.ShortCallerEncoder,
		EncodeName:          nil,
		ConsoleSeparator:    "",
		SkipLineEnding:      false,
		NewReflectedEncoder: nil,
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

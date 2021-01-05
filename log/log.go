package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"sync"
	"time"
)

var _logger *zap.Logger

var lock sync.Mutex

type Env string

type (
	gteDebug struct{} // 大于等于debug级别
	gteInfo  struct{} // 大于等于info级别
	eqWarn   struct{} // 等于warn级别
	gteError struct{} // 大于等于error级别
)

const (
	EnvProduct Env = "product"
	EnvDevelop Env = "develop"
)

func Sync() error {
	return _logger.Sync()
}

func (env gteDebug) Enabled(l zapcore.Level) bool {
	return l >= zapcore.DebugLevel
}

func (env gteDebug) String() string {
	return "debug"
}

func (env gteInfo) Enabled(l zapcore.Level) bool {
	return l >= zapcore.InfoLevel
}

func (env gteInfo) String() string {
	return "info"
}

func (env eqWarn) Enabled(l zapcore.Level) bool {
	return l == zapcore.WarnLevel
}

func (env eqWarn) String() string {
	return "warn"
}

func (env gteError) Enabled(l zapcore.Level) bool {
	return l >= zapcore.ErrorLevel
}

func (env gteError) String() string {
	return "error"
}

func init() {
	_logger = New(EnvDevelop)
}

func New(env Env) *zap.Logger {
	var enablers = make([]zapcore.LevelEnabler, 0, 3)
	var cores = make([]zapcore.Core, 0, 3)
	// 生产环境不输出debug
	switch env {
	case EnvProduct:
		enablers = []zapcore.LevelEnabler{gteInfo{}, eqWarn{}, gteError{}}
	default:
		enablers = []zapcore.LevelEnabler{gteDebug{}, eqWarn{}, gteError{}}
	}
	for i, _ := range enablers {
		cores = append(cores, newCore(enablers[i]))
	}
	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))
	return logger
}

func ReplaceGlobal(logger *zap.Logger) {
	lock.Lock()
	_logger = logger
	lock.Unlock()
}

func Debug(msg string, field ...zap.Field) {
	_logger.Debug(msg, field...)
}

func Info(msg string, field ...zap.Field) {
	_logger.Info(msg, field...)
}

func Warn(msg string, field ...zap.Field) {
	_logger.Warn(msg, field...)
}

func Error(msg string, field ...zap.Field) {
	_logger.Error(msg, field...)
}

func newCore(enabler zapcore.LevelEnabler) zapcore.Core {
	var encoder zapcore.Encoder
	encoder = zapcore.NewConsoleEncoder(newEncoderConfig())
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("./log/%s.log", enabler),
		MaxSize:    100,  // 文件大小，单位：M
		MaxBackups: 50,   // 备份数量
		MaxAge:     365,  // 日志保留天数
		Compress:   true, // 是否压缩
	})
	return zapcore.NewCore(encoder, writer, enabler)
}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func ISO8601TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	encodeTimeLayout(t, "2006-01-02 15:04:05.000", enc)
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

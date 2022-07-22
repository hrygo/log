package log

import (
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 使用注意事项：
// 1. 环境变量 CONF_LOG_TIME_FORMAT 用于设置日期格式，默认为: 2006-01-02T15:04:05.000
// 2. 生产环境日志策略需调用 ProductionDefault 来设置，或者参照此方法根据需要自己修改合适的日志参数
// 3. 使用 ProductionDefault 进行生产环境日志设置时，环境变量 CONF_LOG_PATH 用于设置日志路径，默认为执行程序的当前目录下的logs目录

type Level = zapcore.Level

type Field = zap.Field

var std = New(os.Stdout, DebugLevel, WithCaller(true), zap.AddStacktrace(ErrorLevel))

// ProductionDefault 设置默认生产日志策略
// 参照此方法根据需要自己修改合适的日志参数, 编写自己的初始化方法
func ProductionDefault(opts ...Option) {
	var tops = []TeeOption{
		// 默认JSON格式
		{
			Filename: BasePath() + "all.log",
			Ropt: RotateOptions{
				MaxSize:    100,
				MaxAge:     30,
				MaxBackups: 100,
				Compress:   true,
			},
			Lef: func(lvl Level) bool {
				return lvl <= FatalLevel && lvl > DebugLevel
			},
		},
		// 设置为console格式
		{
			Filename: BasePath() + "error.log",
			Ropt: RotateOptions{
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 10,
				Compress:   false,
				Format:     "console",
			},
			Lef: func(lvl Level) bool {
				return lvl > InfoLevel
			},
		},
	}

	logger := NewTeeWithRotate(tops, opts...)
	ResetDefault(logger)
}

// ResetDefault not safe for concurrent use
func ResetDefault(l *zap.Logger) {
	std = l

	Info = std.Info
	Warn = std.Warn
	Error = std.Error
	DPanic = std.DPanic
	Panic = std.Panic
	Fatal = std.Fatal
	Debug = std.Debug

	Infof = std.Sugar().Infof
	Warnf = std.Sugar().Warnf
	Errorf = std.Sugar().Errorf
	DPanicf = std.Sugar().DPanicf
	Panicf = std.Sugar().Panicf
	Fatalf = std.Sugar().Fatalf
	Debugf = std.Sugar().Debugf
}

type Option = zap.Option

type RotateOptions struct {
	MaxSize    int    // 单个文件最大大小, 单位MB
	MaxAge     int    // 文件最长保存天数
	MaxBackups int    // 最大文件个数
	Compress   bool   // 是否开启压缩
	Format     string // console or json
}

type LevelEnablerFunc func(lvl Level) bool

type TeeOption struct {
	Filename string
	Ropt     RotateOptions
	Lef      LevelEnablerFunc
}

func NewTeeWithRotate(tops []TeeOption, opts ...Option) *zap.Logger {
	var cores []zapcore.Core
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		timeFormat(&t, enc)
	}
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.TimeKey = "created_at"

	for _, top := range tops {
		top := top
		encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
		if top.Ropt.Format == "console" {
			encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
		}
		lef := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return top.Lef(lvl)
		})

		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   top.Filename,
			MaxSize:    top.Ropt.MaxSize,
			MaxBackups: top.Ropt.MaxBackups,
			MaxAge:     top.Ropt.MaxAge,
			Compress:   top.Ropt.Compress,
		})

		core := zapcore.NewCore(encoder, zapcore.AddSync(w), lef)
		cores = append(cores, core)
	}

	return zap.New(zapcore.NewTee(cores...), opts...)
}

// New create a new logger (not support log rotating).
func New(writer io.Writer, level Level, opts ...Option) *zap.Logger {
	if writer == nil {
		panic("the writer is nil")
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		timeFormat(&t, enc)
	}
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(writer),
		level,
	)
	return zap.New(core, opts...)
}

func BasePath() (path string) {
	path = os.Getenv("CONF_LOG_PATH")
	if len(path) == 0 {
		path = "logs/"
		return
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return
}

func Sync() error {
	if std != nil {
		return std.Sync()
	}
	return nil
}

func Default() *zap.Logger {
	return std
}

// 根据环境变量 LOG_TIME_FORMAT 的值来设置日期格式
func timeFormat(t *time.Time, enc zapcore.PrimitiveArrayEncoder) {
	str := os.Getenv("CONF_LOG_TIME_FORMAT")
	if len(str) == 0 {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000"))
	} else {
		enc.AppendString(t.Format(str))
	}
}

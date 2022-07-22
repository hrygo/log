## 使用注意事项：

1. 环境变量 `CONF_LOG_TIME_FORMAT` 用于设置日期格式，默认为: `2006-01-02T15:04:05.000`
2. 生产环境日志策略需调用 `ProductionDefault` 来设置，**或者** 参照此方法根据需要自己修改合适的日志参数
3. 使用 `ProductionDefault` 进行生产环境日志设置时，环境变量 `CONF_LOG_PATH` 用于设置日志路径，默认为执行程序的当前目录下的 `logs` 目录

## 使用示例

```go
package main

import (
	"github.com/ryhgo/log"
	"go.uber.org/zap"
)

func main() {
	defer func() { _ = log.Sync() }() // 确保程序结束前日志flush存储

	hook := zap.Hooks(log.StdoutHooker)        // 添加日志钩子，注意用于异步存储日志到ElasticSearch等日志存储库，需自定义，此处仅为示例
	trace := zap.AddStacktrace(zap.ErrorLevel) // 添加调用栈，Error级别以上会打印
	caller := log.WithCaller(true)             // 添加调用方信息

	// 如果需求与log.ProductionDefault则可以直接调用 log.ProductionDefault(hook, caller, trace) 无需自定义ProductionDefault函数
	ProductionDefault(hook, caller, trace)

	log.Info("test", log.String("hello", "world"))
	log.Infof("hello %s", "world")
}

// ProductionDefault 参照 log.ProductionDefault 自定义的初始化函数
func ProductionDefault(opts ...log.Option) {
	var tops = []log.TeeOption{
		{
			Filename: log.BasePath() + "all.log",
			Ropt: log.RotateOptions{
				MaxSize:    100,
				MaxAge:     30,
				MaxBackups: 100,
				Compress:   true,
			},
			Lef: func(lvl log.Level) bool {
				return lvl <= log.FatalLevel && lvl > log.DebugLevel
			},
		},
		{
			Filename: log.BasePath() + "error.log",
			Ropt: log.RotateOptions{
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 10,
				Compress:   false,
			},
			Lef: func(lvl log.Level) bool {
				return lvl > log.InfoLevel
			},
		},
	}

	logger := log.NewTeeWithRotate(tops, opts...)
	log.ResetDefault(logger)
}
```
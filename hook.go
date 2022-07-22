package log

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// StdoutHooker 系统运行日志钩子函数示例
// 单条日志就是一个结构体格式，本函数拦截每一条日志，您可以进行后续处理，例如：推送到 ElasticSearch 日志库等
func StdoutHooker(entry zapcore.Entry) error {
	// 参数 entry 介绍
	// entry  参数就是单条日志结构体，主要包括字段如下：
	// Level        日志等级
	// Time         当前时间
	// LoggerName   日志名称
	// Message      日志内容
	// Caller       各个文件调用路径
	// Stack        代码调用栈

	// 这里启动一个协程，hook丝毫不会影响程序性能，
	go func(entry zapcore.Entry) {
		// TODO 你可以在这里继续处理系统日志
		fmt.Printf("Stdout Hooker: %#+v\n", entry)
	}(entry)
	return nil
}

package test

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/ryhgo/log"
)

func TestNoneProd(t *testing.T) {
	log.Info("output", log.String("hello", "world"))
	log.Infof("output\thello %s", "world")
}

func TestProductionDefault(t *testing.T) {
	defer func() { _ = log.Sync() }() // 确保程序结束前日志flush存储

	defer func() {
		if err := recover(); err != nil {
			log.Error("recover", log.Reflect("err", err))
		}
	}()

	hook := zap.Hooks(log.StdoutHooker)        // 添加日志钩子
	trace := zap.AddStacktrace(zap.ErrorLevel) // 添加调用栈，Error级别以上会打印
	caller := log.WithCaller(true)             // 添加调用方信息

	log.ProductionDefault(hook, caller, trace)

	log.Debug("", log.String("hello", "world"))
	log.Debugf("hello %s", "world")

	log.Info("", log.String("hello", "world"))
	log.Infof("hello %s", "world")

	log.Warn("", log.String("hello", "world"))
	log.Warnf("hello %s", "world")

	log.Error("", log.String("hello", "world"))
	log.Errorf("hello %s", "world")

	log.Panic("", log.String("hello", "world"))
	log.Panicf("hello %s", "world")

	log.Fatalf("exit -1")

	time.Sleep(time.Minute)
}

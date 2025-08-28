package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/v03413/bepusdt/app"
	"github.com/v03413/bepusdt/app/bot"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
	"github.com/v03413/bepusdt/app/model"
	"github.com/v03413/bepusdt/app/task"
	"github.com/v03413/bepusdt/app/web"
)

type Initializer func() error

var initializers = []Initializer{conf.Init, log.Init, bot.Init, model.Init, task.Init}

func init() {
	fmt.Println("+--------------------------------------------------------------------------------------------------------------+")
	fmt.Println("| 环境变量				设置的值和说明")
	fmt.Println(`| BEPUSDT_DEBUG				"` + os.Getenv("BEPUSDT_DEBUG") + `" 是否开启Debug模式。默认为配置文件中的debug值，支持的值有true/1/false`)
	fmt.Println(`| BEPUSDT_LOG_LEVEL			"` + os.Getenv("BEPUSDT_LOG_LEVEL") + `" 日志等级。默认为debug，支持的值有panic/fatal/error/warn/info/debug/trace`)
	fmt.Println(`| BEPUSDT_LOG_OUTPUT_CONSOLE		"` + os.Getenv("BEPUSDT_LOG_OUTPUT_CONSOLE") + `" 日志是否输出到控制台。默认为false，支持的值有true/1/false`)
	fmt.Println(`| BEPUSDT_EXCHANGE_RATE_UPDATE_INTERVAL "` + os.Getenv("BEPUSDT_EXCHANGE_RATE_UPDATE_INTERVAL") + `" 汇率更新间隔。默认为30m，支持的单位有"m"(分钟), "h"(小时)`)
	fmt.Println("+--------------------------------------------------------------------------------------------------------------+")
	for _, initFunc := range initializers {
		if err := initFunc(); err != nil {

			panic(fmt.Sprintf("初始化失败: %v", err))
		}
	}
}

var COMMIT string

func main() {
	if len(COMMIT) > 0 {
		web.AssetVer = COMMIT
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task.Start(ctx)
	web.Start(ctx)

	fmt.Println("BEpusdt 启动成功，当前版本：" + app.Version)

	{
		var signals = make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)
		<-signals
		cancel()
		runtime.GC()
	}
}

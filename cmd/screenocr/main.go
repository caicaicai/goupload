package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.design/x/hotkey/mainthread"

	"screenocr/internal/config"
	"screenocr/internal/logutil"
	"screenocr/internal/ocr"
	"screenocr/internal/pipeline"
	"screenocr/internal/trigger"
)

func main() {
	// mainthread.Init 在macOS上为全局热键提供必需的主线程事件循环，
	// 在Windows/Linux上则只是直接在主goroutine运行run，不影响行为。
	mainthread.Init(run)
}

func run() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "加载配置失败:", err)
		os.Exit(1)
	}

	logger, closeLog, err := logutil.New(cfg.LogFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化日志失败:", err)
		os.Exit(1)
	}
	defer closeLog()

	recognizer := ocr.New(cfg.OCR.Languages)

	pl := pipeline.New(cfg, recognizer, logger)

	var stopFns []func()

	if cfg.Trigger.Hotkey.Enabled {
		hk, err := trigger.NewHotkeyTrigger(cfg.Trigger.Hotkey.Modifiers, cfg.Trigger.Hotkey.Key)
		if err != nil {
			logger.Error("创建全局热键失败", "error", err)
			os.Exit(1)
		}

		stop, err := hk.Start(func() {
			logger.Info("全局热键触发")
			pl.Run()
		})
		if err != nil {
			logger.Error("注册全局热键失败（macOS请检查是否已授予辅助功能权限）", "error", err)
			os.Exit(1)
		}

		stopFns = append(stopFns, stop)
		logger.Info("全局热键已注册", "modifiers", cfg.Trigger.Hotkey.Modifiers, "key", cfg.Trigger.Hotkey.Key)
	}

	if cfg.Trigger.Timer.Enabled {
		interval, err := time.ParseDuration(cfg.Trigger.Timer.Interval)
		if err != nil {
			logger.Error("定时任务间隔解析失败", "error", err)
			os.Exit(1)
		}

		tm := trigger.NewTimerTrigger(interval)
		stop := tm.Start(func() {
			logger.Info("定时任务触发")
			pl.Run()
		})

		stopFns = append(stopFns, stop)
		logger.Info("定时任务已启动", "interval", cfg.Trigger.Timer.Interval)
	}

	logger.Info("程序已启动，等待触发...", "output_dir", cfg.OutputDir)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("收到退出信号，正在停止...")
	for _, stop := range stopFns {
		stop()
	}
}

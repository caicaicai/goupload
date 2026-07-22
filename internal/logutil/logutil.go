package logutil

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// New 创建一个同时输出到日志文件与标准输出的结构化日志器。
// 若logFile所在目录不存在会自动创建；返回的closeFn用于程序退出前关闭日志文件。
func New(logFile string) (logger *slog.Logger, closeFn func() error, err error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	writer := io.MultiWriter(os.Stdout, f)
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})

	return slog.New(handler), f.Close, nil
}

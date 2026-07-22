package pipeline

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"screenocr/internal/capture"
	"screenocr/internal/config"
	"screenocr/internal/ocr"
)

// Pipeline 串联"截图 -> 保存图片 -> OCR识别 -> 保存文本 -> 记录日志"的整体流程。
type Pipeline struct {
	cfg        *config.Config
	recognizer *ocr.Recognizer
	logger     *slog.Logger
}

// New 创建一个Pipeline实例。
func New(cfg *config.Config, recognizer *ocr.Recognizer, logger *slog.Logger) *Pipeline {
	return &Pipeline{cfg: cfg, recognizer: recognizer, logger: logger}
}

// Run 执行一次完整流程。任何步骤出错都只记录日志，不会向上层panic，
// 以保证热键/定时触发的长期运行不会因单次失败而中断。
func (p *Pipeline) Run() {
	start := time.Now()
	timestamp := start.Format("20060102_150405")

	shots, err := capture.Capture(p.cfg.Monitor)
	if err != nil {
		p.logger.Error("截图失败", "error", err)
		return
	}

	for _, shot := range shots {
		name := timestamp
		if len(shots) > 1 {
			name = fmt.Sprintf("%s_monitor%d", timestamp, shot.MonitorIndex)
		}

		imagePath := filepath.Join(p.cfg.OutputDir, "images", name+".png")
		textPath := filepath.Join(p.cfg.OutputDir, "texts", name+".txt")

		if err := capture.SavePNG(shot.Image, imagePath); err != nil {
			p.logger.Error("保存截图失败", "monitor", shot.MonitorIndex, "error", err)
			continue
		}

		text, err := p.recognizer.Recognize(imagePath, textPath)
		if err != nil {
			p.logger.Error("OCR识别失败", "monitor", shot.MonitorIndex, "image", imagePath, "error", err)
			continue
		}

		p.logger.Info("OCR识别完成",
			"monitor", shot.MonitorIndex,
			"image", imagePath,
			"text_file", textPath,
			"text_length", len([]rune(text)),
			"elapsed", time.Since(start).String(),
		)
	}
}

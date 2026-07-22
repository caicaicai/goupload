package capture

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kbinani/screenshot"
)

// Shot 表示一次截图的结果：截取自哪个显示器索引，以及图像数据。
type Shot struct {
	MonitorIndex int
	Image        *image.RGBA
}

// Capture 按照 monitor 配置（"primary" | "all" | 数字索引字符串）截取屏幕，
// 返回一张或多张截图。
func Capture(monitor string) ([]Shot, error) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return nil, fmt.Errorf("未检测到可用显示器")
	}

	indexes, err := resolveMonitorIndexes(monitor, n)
	if err != nil {
		return nil, err
	}

	shots := make([]Shot, 0, len(indexes))
	for _, idx := range indexes {
		bounds := screenshot.GetDisplayBounds(idx)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return nil, fmt.Errorf("截取显示器%d失败: %w", idx, err)
		}
		shots = append(shots, Shot{MonitorIndex: idx, Image: img})
	}

	return shots, nil
}

func resolveMonitorIndexes(monitor string, count int) ([]int, error) {
	monitor = strings.ToLower(strings.TrimSpace(monitor))

	switch monitor {
	case "", "primary":
		return []int{0}, nil
	case "all":
		indexes := make([]int, count)
		for i := 0; i < count; i++ {
			indexes[i] = i
		}
		return indexes, nil
	default:
		idx, err := strconv.Atoi(monitor)
		if err != nil {
			return nil, fmt.Errorf("非法的monitor配置: %s", monitor)
		}
		if idx < 0 || idx >= count {
			return nil, fmt.Errorf("monitor索引超出范围: %d（当前共有%d个显示器）", idx, count)
		}
		return []int{idx}, nil
	}
}

// SavePNG 将截图保存为PNG文件，若目标目录不存在会自动创建。
func SavePNG(img *image.RGBA, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("编码PNG失败: %w", err)
	}
	return nil
}

//go:build windows

package capture

import (
	"image"

	"github.com/kbinani/screenshot"
)

// platformNumDisplays 返回当前活跃显示器数量。
func platformNumDisplays() int {
	return screenshot.NumActiveDisplays()
}

// platformCaptureDisplay 截取指定索引显示器（0为主屏）的完整画面。
func platformCaptureDisplay(index int) (*image.RGBA, error) {
	bounds := screenshot.GetDisplayBounds(index)
	return screenshot.CaptureRect(bounds)
}

//go:build darwin

package capture

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"unsafe"
)

// 本文件不依赖 github.com/kbinani/screenshot，而是直接基于 CoreGraphics 的
// CGDisplayCreateImage 系列接口自行实现截图。原因：kbinani/screenshot 在其
// darwin实现里根据编译期SDK版本（是否 > 14.4）条件性地链接 ScreenCaptureKit.framework，
// 这在SDK版本较旧（缺少相关头文件/版本宏）的机器上会直接编译失败，且不受
// MACOSX_DEPLOYMENT_TARGET影响。CGDisplayCreateImage是自Mac OS X 10.6起就存在的
// 经典接口，没有这类版本判断问题，兼容性更可靠。

// platformNumDisplays 返回当前活跃显示器数量。
func platformNumDisplays() int {
	var count C.uint32_t
	if C.CGGetActiveDisplayList(0, nil, &count) != C.kCGErrorSuccess {
		return 0
	}
	return int(count)
}

func activeDisplayIDs() ([]C.CGDirectDisplayID, error) {
	n := C.uint32_t(platformNumDisplays())
	if n == 0 {
		return nil, errors.New("未检测到可用显示器")
	}

	ids := make([]C.CGDirectDisplayID, n)
	if C.CGGetActiveDisplayList(n, &ids[0], nil) != C.kCGErrorSuccess {
		return nil, errors.New("获取显示器列表失败")
	}
	return ids, nil
}

// platformCaptureDisplay 截取指定索引显示器（0为主屏）的完整画面。
func platformCaptureDisplay(index int) (*image.RGBA, error) {
	ids, err := activeDisplayIDs()
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(ids) {
		return nil, fmt.Errorf("显示器索引超出范围: %d", index)
	}

	cImg := C.CGDisplayCreateImage(ids[index])
	if unsafe.Pointer(cImg) == nil {
		return nil, errors.New("截取显示器失败：CGDisplayCreateImage返回空，请检查是否已在" +
			"系统设置->隐私与安全性->屏幕录制中为本程序授权")
	}
	defer C.CGImageRelease(cImg)

	width := int(C.CGImageGetWidth(cImg))
	height := int(C.CGImageGetHeight(cImg))
	if width <= 0 || height <= 0 {
		return nil, errors.New("截取显示器失败：图像尺寸非法")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	colorSpace := C.CGColorSpaceCreateWithName(C.kCGColorSpaceSRGB)
	if colorSpace == 0 {
		return nil, errors.New("创建颜色空间失败")
	}
	defer C.CGColorSpaceRelease(colorSpace)

	ctx := C.CGBitmapContextCreate(
		unsafe.Pointer(&img.Pix[0]),
		C.size_t(width),
		C.size_t(height),
		8,
		C.size_t(img.Stride),
		colorSpace,
		C.kCGImageAlphaNoneSkipFirst,
	)
	if ctx == 0 {
		return nil, errors.New("创建绘图上下文失败")
	}
	defer C.CGContextRelease(ctx)

	C.CGContextDrawImage(ctx, C.CGRectMake(0, 0, C.CGFloat(width), C.CGFloat(height)), cImg)

	// CGBitmapContextCreate配合kCGImageAlphaNoneSkipFirst时，内存中每像素4字节排布为
	// [跳过字节, R, G, B]，这里整体左移一位并把Alpha设为255，转换成
	// Go image.RGBA期望的[R, G, B, A]排布。
	for i := 0; i+3 < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = img.Pix[i+1], img.Pix[i+2], img.Pix[i+3], 255
	}

	return img, nil
}

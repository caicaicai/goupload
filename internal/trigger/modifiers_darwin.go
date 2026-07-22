//go:build darwin

package trigger

import "golang.design/x/hotkey"

// altModifier 与 superModifier 的具体常量在不同平台上名称不同
// （Windows为ModAlt/ModWin，macOS为ModOption/ModCmd），因此按平台拆分实现。

func altModifier() hotkey.Modifier {
	return hotkey.ModOption
}

func superModifier() hotkey.Modifier {
	return hotkey.ModCmd
}

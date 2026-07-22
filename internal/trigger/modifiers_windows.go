//go:build windows

package trigger

import "golang.design/x/hotkey"

// altModifier 与 superModifier 的具体常量在不同平台上名称不同
// （Windows为ModAlt/ModWin，macOS为ModOption/ModCmd），因此按平台拆分实现。

func altModifier() hotkey.Modifier {
	return hotkey.ModAlt
}

func superModifier() hotkey.Modifier {
	return hotkey.ModWin
}

// backtickKey 返回Tab键上方"`~"键（数字1左侧）对应的平台按键码。
// golang.design/x/hotkey未提供该键的命名常量，这里直接使用其虚拟键码：
// Windows下为VK_OEM_3（美式键盘布局对应"`~"键）。
func backtickKey() hotkey.Key {
	return hotkey.Key(0xC0)
}

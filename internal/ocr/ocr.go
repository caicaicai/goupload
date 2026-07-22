package ocr

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/zn-chen/sysocr"
)

// Recognizer 封装对操作系统自带OCR能力的调用：
// Windows使用 Windows.Media.Ocr API，macOS使用 Vision Framework。
// 两者均为系统内置能力，程序无需安装或内嵌任何额外的OCR引擎/语言包。
type Recognizer struct {
	languages []string
}

// New 创建一个Recognizer。languages为语言提示（如 "zh-Hans"、"en"），会传递给系统OCR引擎。
//
// 注（Windows）：当前依赖的 sysocr 版本在 Windows 上实际使用系统当前登录用户
// 已安装的OCR语言包（TryCreateFromUserProfileLanguages），languages 配置暂不生效；
// 在 macOS 上则会正确传递给 Vision Framework 作为语言提示。
func New(languages []string) *Recognizer {
	return &Recognizer{languages: languages}
}

// Recognize 对指定图片文件进行OCR识别，识别结果会写入outputTextPath，
// 同时返回识别出的文本内容，供上层记录日志使用。
func (r *Recognizer) Recognize(imagePath, outputTextPath string) (string, error) {
	result, err := sysocr.Recognize(sysocr.Options{
		Input:     sysocr.Input{FilePath: imagePath},
		Languages: r.languages,
	})
	if err != nil {
		return "", fmt.Errorf(
			"系统OCR识别失败: %w（Windows需系统已安装对应语言的OCR语言包，可在设置->时间和语言->语言中添加；"+
				"macOS需10.15及以上版本）",
			err,
		)
	}

	text := collapseCJKSpaces(result.Text)

	if err := os.MkdirAll(filepath.Dir(outputTextPath), 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	if err := os.WriteFile(outputTextPath, []byte(text), 0o644); err != nil {
		return "", fmt.Errorf("保存OCR结果文件失败: %w", err)
	}

	return text, nil
}

// collapseCJKSpaces 去除中日韩(CJK)文字之间被系统OCR引擎错误插入的空格。
// Windows.Media.Ocr 在识别一整行连续中文时，常将每个汉字识别为独立的"单词"，
// 并在拼接行文本时于其间插入空格，而中文本身并不使用空格分词，因此需在此纠正。
func collapseCJKSpaces(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes))

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ' ' && len(out) > 0 && isCJK(out[len(out)-1]) {
			j := i
			for j < len(runes) && runes[j] == ' ' {
				j++
			}
			if j < len(runes) && isCJK(runes[j]) {
				i = j - 1
				continue
			}
		}
		out = append(out, r)
	}

	return string(out)
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

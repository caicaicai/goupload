package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 是程序的完整配置结构，对应 config.yaml 文件内容。
type Config struct {
	OutputDir string        `yaml:"output_dir"`
	Monitor   string        `yaml:"monitor"`
	OCR       OCRConfig     `yaml:"ocr"`
	Trigger   TriggerConfig `yaml:"trigger"`
	LogFile   string        `yaml:"log_file"`
}

// OCRConfig 配置系统自带OCR引擎（Windows: Windows.Media.Ocr；macOS: Vision Framework）的语言提示。
type OCRConfig struct {
	Languages []string `yaml:"languages"`
}

type TriggerConfig struct {
	Hotkey HotkeyConfig `yaml:"hotkey"`
	Timer  TimerConfig  `yaml:"timer"`
}

type HotkeyConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Modifiers []string `yaml:"modifiers"`
	Key       string   `yaml:"key"`
}

type TimerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}

// Default 返回一份带有合理默认值的配置。
func Default() *Config {
	return &Config{
		OutputDir: "./output",
		Monitor:   "primary",
		OCR: OCRConfig{
			Languages: []string{"zh-Hans", "en"},
		},
		Trigger: TriggerConfig{
			Hotkey: HotkeyConfig{
				Enabled:   true,
				Modifiers: []string{},
				Key:       "`",
			},
			Timer: TimerConfig{
				Enabled:  false,
				Interval: "30m",
			},
		},
		LogFile: "./logs/app.log",
	}
}

// Load 从指定路径加载YAML配置文件，缺失字段使用默认值填充，并做基本校验。
// 若path不存在，则直接返回默认配置。
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}

	return cfg, nil
}

// applyDefaults 对YAML中未显式配置（解析后为空值）的字段补充默认值。
func (c *Config) applyDefaults() {
	def := Default()

	if strings.TrimSpace(c.OutputDir) == "" {
		c.OutputDir = def.OutputDir
	}
	if strings.TrimSpace(c.Monitor) == "" {
		c.Monitor = def.Monitor
	}
	if len(c.OCR.Languages) == 0 {
		c.OCR.Languages = def.OCR.Languages
	}
	if strings.TrimSpace(c.LogFile) == "" {
		c.LogFile = def.LogFile
	}
	if c.Trigger.Hotkey.Enabled && len(c.Trigger.Hotkey.Modifiers) == 0 && strings.TrimSpace(c.Trigger.Hotkey.Key) == "" {
		c.Trigger.Hotkey = def.Trigger.Hotkey
	}
	if c.Trigger.Timer.Enabled && strings.TrimSpace(c.Trigger.Timer.Interval) == "" {
		c.Trigger.Timer.Interval = def.Trigger.Timer.Interval
	}
}

// Validate 校验配置字段的合法性。
func (c *Config) Validate() error {
	monitor := strings.ToLower(strings.TrimSpace(c.Monitor))
	if monitor != "primary" && monitor != "all" {
		if _, err := parseMonitorIndex(monitor); err != nil {
			return fmt.Errorf("monitor 配置非法，应为 primary/all/数字索引，实际为: %s", c.Monitor)
		}
	}

	if len(c.OCR.Languages) == 0 {
		return fmt.Errorf("ocr.languages 不能为空")
	}

	if !c.Trigger.Hotkey.Enabled && !c.Trigger.Timer.Enabled {
		return fmt.Errorf("trigger.hotkey 与 trigger.timer 至少需要启用一个")
	}

	if c.Trigger.Hotkey.Enabled {
		if strings.TrimSpace(c.Trigger.Hotkey.Key) == "" {
			return fmt.Errorf("trigger.hotkey.enabled 为 true 时，key 不能为空")
		}
	}

	if c.Trigger.Timer.Enabled {
		if _, err := time.ParseDuration(c.Trigger.Timer.Interval); err != nil {
			return fmt.Errorf("trigger.timer.interval 格式非法（示例: 30m, 1h）: %w", err)
		}
	}

	return nil
}

// MonitorIndex 返回 monitor 配置解析出的具体屏幕索引（仅当 monitor 为数字索引时有意义）。
func parseMonitorIndex(s string) (int, error) {
	var idx int
	_, err := fmt.Sscanf(s, "%d", &idx)
	if err != nil {
		return 0, err
	}
	return idx, nil
}

# screenocr

一个无界面的桌面后台工具：通过**全局快捷键**或**定时任务**触发全屏截图，调用**操作系统自带的OCR能力**进行中英文识别，并将截图与识别文本保存到本地文件。支持 Windows 与 macOS，**无需安装或内嵌任何第三方OCR引擎**。

## 功能特点

- 无GUI，作为后台进程运行
- 全局快捷键触发（如 `Ctrl+Shift+S`）和/或定时任务触发（如每30分钟一次），两者可同时启用
- 全屏截图，支持仅截主屏、指定屏幕或所有屏幕
- OCR识别直接调用操作系统内置引擎（Windows: `Windows.Media.Ocr`；macOS: `Vision Framework`），**编译完成即可直接运行，不需要用户额外安装Tesseract等任何软件**
- 每次识别结果均保存为 PNG 截图 + 同名 TXT 文本，并写入日志文件

## 目录结构

```
.
├── cmd/screenocr/main.go        程序入口
├── internal/config              配置加载与校验
├── internal/capture             截图（Windows基于kbinani/screenshot；macOS自行实现CoreGraphics调用）
├── internal/ocr                 OCR识别（基于 zn-chen/sysocr 调用系统OCR API）
├── internal/trigger             全局热键 / 定时任务触发器
├── internal/pipeline            截图->OCR->保存->日志 主流程编排
├── internal/logutil             文件日志封装
└── config.example.yaml          配置文件示例
```

## 运行环境要求

不需要安装Tesseract或其它OCR软件，只需满足系统OCR引擎本身的最低版本要求：

- **Windows 10 及以上**：使用系统内置的 `Windows.Media.Ocr` API。中文识别依赖系统已安装的"OCR语言包"，中文版Windows通常已自带；纯英文系统若识别中文效果不佳，可在 **设置 → 时间和语言 → 语言和区域** 中为中文添加"手写和键入语言"对应的语言包（一次性、系统级操作，与本程序无关）。
- **macOS 10.15（Catalina）及以上**：使用系统内置的 `Vision Framework`，开箱即用，通常无需额外配置。

首次运行时，各平台仍可能出现如下系统权限弹窗，需要用户手动允许：

- **macOS**：
  - **屏幕录制**（Screen Recording）：截图功能依赖，位于 系统设置 → 隐私与安全性 → 屏幕录制
  - **辅助功能**（Accessibility）：全局快捷键监听依赖，位于 系统设置 → 隐私与安全性 → 辅助功能

## 配置

复制 `config.example.yaml` 为 `config.yaml`，按需修改：

```yaml
output_dir: "./output"
monitor: "primary"

ocr:
  languages: ["zh-Hans", "en"]

trigger:
  hotkey:
    enabled: true
    modifiers: ["ctrl", "shift"]
    key: "S"
  timer:
    enabled: false
    interval: "30m"

log_file: "./logs/app.log"
```

字段说明见 `config.example.yaml` 中的注释。`trigger.hotkey` 与 `trigger.timer` 至少需启用一个。

## 构建与运行

截图、全局热键、OCR在macOS上均依赖CGO（调用Cocoa/CoreGraphics/Vision Framework），因此需要在对应平台上分别原生构建（Windows端则完全不需要CGO/C编译器）：

```bash
# 在目标平台（Windows 或 macOS）上执行
go build -o screenocr ./cmd/screenocr

# 运行（默认读取当前目录下的config.yaml）
./screenocr

# 或指定配置文件路径
./screenocr -config /path/to/config.yaml
```

程序启动后会持续在后台运行，按 `Ctrl+C` 或发送终止信号可优雅退出。

## 跨平台构建（GitHub Actions）

由于macOS上的截图/热键/OCR均依赖CGO，无法在Windows上直接交叉编译，仓库内已提供 [.github/workflows/build.yml](.github/workflows/build.yml)，推送到GitHub后会自动在对应平台的官方Runner上原生构建：

| Runner | 产物 |
| --- | --- |
| `windows-latest` | `screenocr-windows-amd64` |
| `macos-14`（Apple Silicon） | `screenocr-macos-arm64` |
| `macos-15-intel`（Intel） | `screenocr-macos-amd64` |

使用方式：

1. 将本仓库推送到GitHub（`git init` → 关联远程仓库 → `git push`）。
2. 推送到 `main` 分支或提交PR会自动触发构建（仅验证编译，产物可在Actions运行详情的 **Artifacts** 里下载）。
3. 打一个 `v` 开头的tag（如 `git tag v0.1.0 && git push origin v0.1.0`）会额外触发创建GitHub Release，并把三个平台的构建产物打包附加到Release上。
4. 也可以在Actions页面手动触发（`workflow_dispatch`），无需等推送。

## 输出说明

- 截图：`output_dir/images/<时间戳>.png`（多屏时为 `<时间戳>_monitor<索引>.png`）
- 识别文本：`output_dir/texts/<时间戳>.txt`（与截图同名）
- 运行日志：`log_file` 指定的文件，记录每次触发的耗时、结果与错误信息

## 常见问题

- **macOS下热键无响应/注册报错**：检查是否已在"辅助功能"中为本程序授权。
- **macOS下截图为黑屏**：检查是否已在"屏幕录制"中为本程序授权。
- **中文识别效果不理想/识别为空**：确认系统已安装对应语言的OCR语言包（Windows可在系统设置的语言选项中添加；macOS一般自带多语言支持），并确认 `ocr.languages` 中包含了正确的语言代码（如 `zh-Hans`）。
- **需要识别其它语言**：在 `ocr.languages` 中添加对应的语言代码（如日语 `ja`、韩语 `ko`），具体支持的语言集合以操作系统OCR组件为准。
- **Windows下修改`ocr.languages`似乎不生效**：当前依赖的系统OCR封装库在Windows上使用的是系统当前用户已安装的OCR语言包（而非`ocr.languages`配置），因此实际识别语言取决于系统"设置→时间和语言→语言"中已安装的语言包；`ocr.languages`在macOS（Vision Framework）上会正确生效。
- **macOS下编译报错 `ScreenCaptureKit/ScreenCaptureKit.h file not found`，或启动报错 `Library not loaded: .../ScreenCaptureKit`**：这是历史遗留问题，早期版本的macOS截图实现依赖第三方库`kbinani/screenshot`，其内部会根据编译期SDK版本条件性链接`ScreenCaptureKit.framework`，在SDK版本较旧或系统版本低于12.3（Monterey）的机器上会编译或运行失败。当前版本已在 [internal/capture/platform_darwin.go](internal/capture/platform_darwin.go) 中改为自行实现（仅调用自Mac OS X 10.6起就存在的经典`CGDisplayCreateImage`接口），不再有此问题。如果你仍遇到这个报错，请确认代码已更新到最新版本（`git pull`）后重新构建。

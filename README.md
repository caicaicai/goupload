# screenocr

一个无界面的桌面后台工具：通过**全局快捷键**或**定时任务**触发全屏截图，调用**操作系统自带的OCR能力**进行中英文识别，并将截图与识别文本保存到本地文件。支持 Windows 与 macOS，**无需安装或内嵌任何第三方OCR引擎**。

## 功能特点

- 无GUI，作为后台进程运行
- 全局快捷键触发（默认是Tab键上方的波浪键 `` ` ``，单独按一下即可，也可自行改成组合键）和/或定时任务触发（如每30分钟一次），两者可同时启用
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

  由于`screenocr`是一个未签名的原始可执行文件（不是`.app`应用包），首次运行时系统通常不会自动弹出授权对话框，需要手动添加：进入 **系统设置 → 隐私与安全性 → 辅助功能**（截图权限对应"屏幕录制"页面，步骤相同），点击左下角 **"+"**，在文件选择框中找到编译出的`screenocr`可执行文件并添加，确保开关处于打开状态，再重新运行程序。
  每次`go build`重新编译后，二进制文件会变化，macOS有时会认为这是"新程序"而清除已有授权，如果重新编译后又提示权限报错，需要重复上述步骤重新添加一次（或先移除列表中旧的条目再重新添加）。

  **重要**：如果是从终端（Terminal.app / iTerm2等）直接运行`screenocr`（而不是把它做成独立的`.app`应用），仅添加`screenocr`本身通常不够——macOS的隐私权限系统（TCC）经常把命令行工具的权限请求归属到**启动它的终端应用**上。如果加了`screenocr`之后仍然报同样的权限错误，请额外把你正在使用的终端应用（如"终端.app"，路径通常是`应用程序 → 实用工具 → 终端`；iTerm2则是`应用程序 → iTerm`）也添加到"辅助功能"（以及需要截图时的"屏幕录制"）列表里，并**完全退出该终端应用（`Cmd+Q`）后重新打开**（仅关闭窗口不会让新权限生效），再重新运行程序。

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
    modifiers: []
    key: "`"
  timer:
    enabled: false
    interval: "30m"

log_file: "./logs/app.log"
```

字段说明见 `config.example.yaml` 中的注释。`trigger.hotkey` 与 `trigger.timer` 至少需启用一个。

`trigger.hotkey.key` 支持：单个字母/数字（如 `"S"`、`"9"`）、`` "`" ``（Tab键上方的波浪键）、`F1`~`F20`、`SPACE`/`TAB`/`ESC`/`ENTER`/`DELETE`/方向键等。`modifiers` 可以是 `["ctrl", "shift", "alt", "win"（macOS上等价于Cmd）]` 的任意组合，留空 `[]` 表示不需要按修饰键，单独按`key`即可触发。

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

- **macOS下热键无响应，或报错 `hotkey: failed to register, grant the application Accessibility (Input Monitoring) permission`**：说明还没有在"辅助功能"中为本程序授权，具体步骤见上方"运行环境要求"章节。
- **macOS下截图为黑屏，或截到的是一片空白桌面（没有任何窗口/图标内容）**：没有报错，但这正是没有真正获得"屏幕录制"权限时的典型表现——`CGDisplayCreateImage`在无权限时不会报错，而是静默返回一张空的桌面背景图。请确认"屏幕录制"里加的是**终端应用本身**（而不只是`screenocr`可执行文件，原因见上方"运行环境要求"章节的说明），加完后同样需要**完全退出终端（`Cmd+Q`）再重新打开**才会生效。
- **中文识别效果不理想/识别为空**：确认系统已安装对应语言的OCR语言包（Windows可在系统设置的语言选项中添加；macOS一般自带多语言支持），并确认 `ocr.languages` 中包含了正确的语言代码（如 `zh-Hans`）。
- **需要识别其它语言**：在 `ocr.languages` 中添加对应的语言代码（如日语 `ja`、韩语 `ko`），具体支持的语言集合以操作系统OCR组件为准。
- **Windows下修改`ocr.languages`似乎不生效**：当前依赖的系统OCR封装库在Windows上使用的是系统当前用户已安装的OCR语言包（而非`ocr.languages`配置），因此实际识别语言取决于系统"设置→时间和语言→语言"中已安装的语言包；`ocr.languages`在macOS（Vision Framework）上会正确生效。
- **macOS下编译报错 `ScreenCaptureKit/ScreenCaptureKit.h file not found`，或启动报错 `Library not loaded: .../ScreenCaptureKit`**：这是历史遗留问题，早期版本的macOS截图实现依赖第三方库`kbinani/screenshot`，其内部会根据编译期SDK版本条件性链接`ScreenCaptureKit.framework`，在SDK版本较旧或系统版本低于12.3（Monterey）的机器上会编译或运行失败。当前版本已在 [internal/capture/platform_darwin.go](internal/capture/platform_darwin.go) 中改为自行实现（仅调用自Mac OS X 10.6起就存在的经典`CGDisplayCreateImage`接口），不再有此问题。如果你仍遇到这个报错，请确认代码已更新到最新版本（`git pull`）后重新构建。

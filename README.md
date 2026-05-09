# ncmdump-go

基于 https://github.com/taurusxin/ncmdump 的 Golang 移植版

支持网易云音乐最新的 3.x 版本，但需要注意：从 3. x开始的某些网易云音乐版本不再在 ncm 文件中内置封面图片，本项目支持从网易服务器上自动下载对应歌曲的封面图并写入到最终的音乐文件中

你也可以去 https://git.taurusxin.com/taurusxin/ncmdump-gui 下载基于本项目的 gui 可视化图形应用，只需简单点击即可自动转换。

## 如何提 Issue

由于本站恶意机器人注册过多，已关闭账号注册，如果需要提 Issue 请前往 [GitHub](https://github.com/taurusxin/ncmdump)，必须注明 Issue 的主题为 ncmdump-go，敬请谅解。

## 安装

你可以使用去 [releases](https://git.taurusxin.com/taurusxin/ncmdump-go/releases/latest) 下载最新版预编译好的二进制文件，或者你也可以用包管理器来安装

```shell
# Windows Scoop
scoop bucket add taurusxin https://git.taurusxin.com/taurusxin/scoop-bucket.git  # 添加 scoop 源
scoop install ncmdump-go # 安装 ncmdump-go

# macOS & Linux 之后会支持
```

## 使用方法

## 图形界面（Fyne）

启动内置的桌面图形界面（添加 .ncm 或整文件夹、可选输出目录与递归扫描，与命令行相同的转换与元数据逻辑）：

```shell
ncmdump-go --gui
```

### 打包为 macOS 应用程序（.app）

1. 安装 Fyne 命令行工具（任选其一）：

   ```shell
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

2. 在项目根目录执行：

   ```shell
   ./package-macos-app.sh
   ```

   会生成 `build/ncmdump-go.app`（需本机已配置 Xcode Command Line Tools，且 **CGO 已启用**，以便链接图形界面）。

3. 安装到「应用程序」文件夹，例如在终端执行：

   ```shell
   cp -R build/ncmdump-go.app /Applications/
   ```

   也可在访达中把 `build/ncmdump-go.app` 拖入「应用程序」。首次打开未签名应用若被拦截，可在「系统设置 → 隐私与安全性」中允许，或对应用图标使用右键 **打开**。

   **说明**：从访达双击 `.app` 时，系统常只传入可执行文件路径（无 `-psn_` 等附加参数），若仍按「无参 CLI」逻辑会直接退出，看起来像闪退。当前版本会检测是否运行于 `*.app/Contents/MacOS/` 下并在无其它参数时启动图形界面。若在终端执行包内二进制并需要命令行帮助，请显式加上 `-h` 等参数。

使用 `-h` 或 `--help` 参数来打印帮助

```shell
ncmdump-go -h
```

使用 `-v` 或 `--version` 参数来打印版本信息

```shell
ncmdump-go -v
```

处理单个或多个文件

```shell
ncmdump-go 1.ncm 2.ncm...
```

使用 `-d` 参数来指定一个文件夹，对文件夹下的所有以 ncm 为扩展名的文件进行批量处理

```shell
ncmdump-go -d source_dir
```

使用 `-r` 配合 `-d` 参数来递归处理文件夹下的所有以 ncm 为扩展名的文件

```shell
ncmdump-go -d source_dir -r
```

使用 `-o` 参数来指定输出目录，将转换后的文件输出到指定目录，该参数支持与 `-r` 参数一起使用

```shell
# 处理单个或多个文件并输出到指定目录
ncmdump-go 1.ncm 2.ncm -o output_dir

# 处理文件夹下的所有以 ncm 为扩展名并输出到指定目录，不包含子文件夹
ncmdump-go -d source_dir -o output_dir

# 递归处理文件夹并输出到指定目录，并保留目录结构
ncmdump-go -d source_dir -o output_dir -r
```

## 开发

使用 go module 下载 ncmdump-go 包

```shell
go get -u git.taurusxin.com/taurusxin/ncmdump-go
```

导入并使用

```go
package main

import (
	"fmt"
	"git.taurusxin.com/taurusxin/ncmdump-go/ncmcrypt"
)

func main() {
	filePath := "test.ncm"
	
	// 创建实例
	ncm, err := ncmcrypt.NewNeteaseCloudMusic(filePath)
	if err != nil {
		fmt.Printf("Reading '%s' failed: '%s'\n", filePath, err.Error())
		return
	}
	
	// 转换格式，若目标文件夹为空，则保存在原目录
	dumpResult, err := ncm.Dump("")
	if err != nil {
		fmt.Printf("Processing '%s' failed: '%s'\n", filePath, err.Error())
	}
	if dumpResult {
		// 使用源文件的元数据修补转换后的音乐文件
		// 注意：自网易云音乐 3.0 版本开始，ncm 文件中不再内嵌专辑封面图片，参数若为 true 则表示从网易服务器上下载图片并嵌入到目标音乐文件（需要联网）
		metadata, err := ncm.FixMetadata(true)
		if !metadata {
			fmt.Printf("Fix metadata for '%s' failed: '%s'", filePath, err.Error())
		}
		fmt.Printf("'%s' -> '%s'\n", filePath, ncm.GetDumpFilePath())
	}
}
```


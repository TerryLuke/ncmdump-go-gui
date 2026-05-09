#!/usr/bin/env bash
# 在 macOS 上打包为可拖入「应用程序」的 .app bundle（Fyne GUI）。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

command -v fyne >/dev/null 2>&1 || {
  echo "未找到 fyne 命令。请安装官方新版 CLI：" >&2
  echo "  go install fyne.io/tools/cmd/fyne@latest" >&2
  exit 1
}

mkdir -p build
rm -rf build/ncmdump-go.app

if [[ ! -f Icon.png ]]; then
  echo "缺少 Icon.png（应用图标），无法打包。" >&2
  exit 1
fi

# 从 FyneApp.toml 同步版本较繁琐，此处与 internal/version 保持一致即可
APP_VER=$(grep -E '^const String' internal/version/version.go | sed -n 's/.*"\(.*\)".*/\1/p')

# fyne.io/tools/cmd/fyne 使用 --id / --app-version（旧 fyne.io/fyne/v2/cmd/fyne 已弃用）
fyne package --release \
  --id com.my.ncmdump-go \
  --name "ncmdump-go" \
  --app-version "${APP_VER}" \
  --icon Icon.png

mv -f ncmdump-go.app build/

echo ""
echo "已生成: $ROOT/build/ncmdump-go.app"
echo "安装到应用程序文件夹："
echo "  cp -R \"$ROOT/build/ncmdump-go.app\" /Applications/"
echo "或拖动 build/ncmdump-go.app 到「访达 → 应用程序」。"

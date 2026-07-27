#!/usr/bin/env bash
# WeChat Mini Program — requires WeChat DevTools
# Usage: ./start.sh [open|build]

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

case "${1:-open}" in
  open)
    echo "请手动打开微信开发者工具，导入 $SCRIPT_DIR/apps/miniprogram 目录"
    echo "项目路径: $SCRIPT_DIR/apps/miniprogram"
    ;;
  build)
    echo "微信小程序构建需要微信开发者工具命令行工具 (cli)"
    echo "安装: npm install -g @wechat/miniprogram-cli"
    echo "运行: cd $SCRIPT_DIR/apps/miniprogram && wechat miniprogram build"
    ;;
  *)
    echo "用法: $0 [open|build]"
    echo "  open  - 显示手动打开步骤"
    echo "  build - 显示构建说明"
    ;;
esac

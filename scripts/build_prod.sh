#!/usr/bin/env bash
# 构建 admin / web_app，拷贝 dist 到 allinone embed 目录，再编译单进程二进制。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ADMIN_DIST="$ROOT/admin/dist"
WEB_DIST="$ROOT/web_app/dist"
EMBED_ADMIN="$ROOT/server/cmd/allinone/static/admin"
EMBED_WEB="$ROOT/server/cmd/allinone/static/web_app"
OUT_DIR="$ROOT/bin"
OUT_NAME="mall"

if [[ "$(uname -s)" == MINGW* || "$(uname -s)" == MSYS* || "$(uname -s)" == CYGWIN* ]]; then
  OUT_NAME="mall.exe"
fi

echo "==> build admin"
cd "$ROOT/admin"
if [[ -f package-lock.json ]]; then
  npm ci
else
  npm install
fi
npm run build

echo "==> build web_app"
cd "$ROOT/web_app"
if [[ -f package-lock.json ]]; then
  npm ci
else
  npm install
fi
npm run build

echo "==> copy dist into embed dirs"
rm -rf "$EMBED_ADMIN" "$EMBED_WEB"
mkdir -p "$EMBED_ADMIN" "$EMBED_WEB"
cp -a "$ADMIN_DIST/." "$EMBED_ADMIN/"
cp -a "$WEB_DIST/." "$EMBED_WEB/"
# 保证目录可被 git 跟踪占位（构建产物本身仍被 gitignore）
touch "$EMBED_ADMIN/.gitkeep" "$EMBED_WEB/.gitkeep"

echo "==> go build allinone"
mkdir -p "$OUT_DIR"
cd "$ROOT/server"
CGO_ENABLED=0 go build -o "$OUT_DIR/$OUT_NAME" ./cmd/allinone

echo "==> done: $OUT_DIR/$OUT_NAME"
echo "Run from server/ (or adjust config paths), e.g.:"
echo "  cd \"$ROOT/server\" && \"../bin/$OUT_NAME\" -conf configs/config_prod.yaml"

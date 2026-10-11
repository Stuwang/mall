#!/usr/bin/env bash
# 构建 admin / web_app，拷贝 dist 到 allinone embed 目录，再编译单进程二进制。
#
# 用法:
#   sh scripts/build_prod.sh              # 当前平台
#   sh scripts/build_prod.sh linux        # linux/amd64（默认 arch）
#   sh scripts/build_prod.sh linux arm64  # linux/arm64
#   sh scripts/build_prod.sh windows amd64
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ADMIN_DIST="$ROOT/admin/dist"
WEB_DIST="$ROOT/web_app/dist"
EMBED_ADMIN="$ROOT/server/cmd/allinone/static/admin"
EMBED_WEB="$ROOT/server/cmd/allinone/static/web_app"
OUT_DIR="$ROOT/bin"

TARGET_OS="${1:-$(go env GOOS)}"
TARGET_ARCH="${2:-}"
if [[ -z "$TARGET_ARCH" ]]; then
  if [[ "$TARGET_OS" == "$(go env GOOS)" ]]; then
    TARGET_ARCH="$(go env GOARCH)"
  else
    TARGET_ARCH="amd64"
  fi
fi

OUT_NAME="mall-${TARGET_OS}-${TARGET_ARCH}"
if [[ "$TARGET_OS" == "windows" ]]; then
  OUT_NAME="${OUT_NAME}.exe"
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
touch "$EMBED_ADMIN/.gitkeep" "$EMBED_WEB/.gitkeep"

echo "==> go build allinone (GOOS=$TARGET_OS GOARCH=$TARGET_ARCH)"
mkdir -p "$OUT_DIR"
cd "$ROOT/server"
CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -o "$OUT_DIR/$OUT_NAME" ./cmd/allinone

# 当前平台额外写一份短名，方便本地直接跑
HOST_OS="$(go env GOOS)"
HOST_ARCH="$(go env GOARCH)"
if [[ "$TARGET_OS" == "$HOST_OS" && "$TARGET_ARCH" == "$HOST_ARCH" ]]; then
  SHORT_NAME="mall"
  if [[ "$TARGET_OS" == "windows" ]]; then
    SHORT_NAME="mall.exe"
  fi
  cp -f "$OUT_DIR/$OUT_NAME" "$OUT_DIR/$SHORT_NAME"
  echo "==> also: $OUT_DIR/$SHORT_NAME"
fi

echo "==> done: $OUT_DIR/$OUT_NAME"
echo "Run from server/ (or adjust config paths), e.g.:"
if [[ "$TARGET_OS" == "linux" ]]; then
  echo "  scp bin/$OUT_NAME user@host:/opt/mall/"
  echo "  cd /opt/mall && ./$OUT_NAME -conf configs/config_prod.yaml"
else
  echo "  cd \"$ROOT/server\" && \"../bin/$OUT_NAME\" -conf configs/config_prod.yaml"
fi

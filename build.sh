#!/usr/bin/env bash
#
# 配布用のバイナリと、その隣に置くアクセスコード一覧を dist/ に用意する
# (管理者が手元で実行する)。
#
#   ./build.sh                  # access_codes.txt を dist/ に添える
#   ./build.sh path/to/list.txt # 一覧の場所を指定する
#   ./build.sh --no-codes       # バイナリだけ作る (一覧は添えない)
#
# 実行ファイルは同じフォルダの access_codes.txt を読むので、配るときは
# 必ず 2 つを同じ場所に置く。一覧はアクセスコードそのものなので、
# ラボ内からしか見えない場所に置くこと。
set -euo pipefail

cd "$(dirname "$0")"

codes_file="access_codes.txt"
with_codes=1
if [ "${1:-}" = "--no-codes" ]; then
  with_codes=0
elif [ -n "${1:-}" ]; then
  codes_file="$1"
fi

if [ "$with_codes" = 1 ] && [ ! -f "$codes_file" ]; then
  echo "一覧ファイルがありません: $codes_file" >&2
  echo "1 行に 1 台「シリアル番号 アクセスコード」を並べたファイルを用意してください。" >&2
  exit 1
fi

go test ./...

# 配る前に、一覧が想定どおり読めるかを実際に動かして確かめる。
# ここで件数が合わなければ、一覧の書式が崩れている。
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
printf '{\n    "user_access_code": {}\n}\n' > "$tmp/check.conf"
go build -trimpath -o "$tmp/check" .
if [ "$with_codes" = 1 ]; then
  echo "--- 配る一覧の確認 ---"
  "$tmp/check" -n -f "$codes_file" "$tmp/check.conf" < /dev/null
  echo "----------------------"
fi

mkdir -p dist
build() { # GOOS GOARCH 出力名
  echo "building dist/bambu_access_code-$3"
  GOOS="$1" GOARCH="$2" go build -trimpath -o "dist/bambu_access_code-$3" .
}

build windows amd64 windows-x64.exe
build windows arm64 windows-arm64.exe
build windows 386   windows-x86.exe
build darwin  arm64 macos-arm64
build darwin  amd64 macos-x64
build linux   amd64 linux-x86_64

if [ "$with_codes" = 1 ]; then
  cp "$codes_file" dist/access_codes.txt
  chmod 600 dist/access_codes.txt
fi

echo
echo "できあがり:"
ls -lh dist/
if [ "$with_codes" = 1 ]; then
  echo
  echo "配るときは実行ファイルと dist/access_codes.txt を同じ場所に置く。"
fi

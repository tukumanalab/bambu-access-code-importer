#!/usr/bin/env bash
#
# アクセスコードを埋め込んだ配布用バイナリを作る (管理者が手元で実行する)。
#
#   ./build.sh                  # access_codes.txt を埋め込む
#   ./build.sh path/to/list.txt # 一覧の場所を指定する
#   ./build.sh --no-codes       # 埋め込まず、実行時に貼り付ける版を作る
#
# できたバイナリはラボ内からしか見えない場所に置く。埋め込んだコードは
# strings コマンドで取り出せるので、公開の場所には絶対に置かないこと。
set -euo pipefail

cd "$(dirname "$0")"

codes_file="access_codes.txt"
embed=1
if [ "${1:-}" = "--no-codes" ]; then
  embed=0
elif [ -n "${1:-}" ]; then
  codes_file="$1"
fi

ldflags=""
if [ "$embed" = 1 ]; then
  if [ ! -f "$codes_file" ]; then
    echo "一覧ファイルがありません: $codes_file" >&2
    echo "1 行に 1 台「シリアル番号 アクセスコード」を並べたファイルを用意してください。" >&2
    exit 1
  fi
  # base64 にすると空白・改行が消えるので -X にそのまま渡せる。
  b64=$(base64 < "$codes_file" | tr -d '\n')
  ldflags="-X main.embeddedCodesB64=$b64"
fi

go test ./...

# 配る前に、埋め込んだ一覧が想定どおり読めるかを実際に動かして確かめる。
# ここで件数が合わなければ、一覧の書式が崩れている。
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
printf '{\n    "user_access_code": {}\n}\n' > "$tmp/check.conf"
go build -trimpath -ldflags "$ldflags" -o "$tmp/check" .
if [ "$embed" = 1 ]; then
  echo "--- 埋め込んだ一覧の確認 ---"
  "$tmp/check" -n "$tmp/check.conf" < /dev/null
  echo "---------------------------"
fi

mkdir -p dist
build() { # GOOS GOARCH 出力名
  echo "building dist/patch_access_code-$3"
  GOOS="$1" GOARCH="$2" go build -trimpath -ldflags "$ldflags" \
    -o "dist/patch_access_code-$3" .
}

build windows amd64 windows-x64.exe
build windows arm64 windows-arm64.exe
build windows 386   windows-x86.exe
build darwin  arm64 macos-arm64
build darwin  amd64 macos-x64
build linux   amd64 linux-x86_64

echo
echo "できあがり:"
ls -lh dist/

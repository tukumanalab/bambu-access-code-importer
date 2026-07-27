# 開発者向けメモ

使い方は [README.md](README.md) を参照。ここにはビルド・リリースと
実装上の判断を書いておく。

## リポジトリ構成

| ファイル | 用途 |
|---|---|
| `patch_access_code.rb` | 書き戻し処理の本体。spinel でバイナリ化する |
| `.github/workflows/build.yml` | タグ push でビルドして Releases に添付 |

アクセスコードはバイナリにもソースにも含まれない（実行時に貼り付けるか
`-f` で渡す）。手元に置く `access_codes.txt` は `.gitignore` で除外済み
（廃止した `access_codes.json` も、古い手元コピーが残っていても
コミットされないよう除外したままにしてある）。

## リリース手順

タグを push すると GitHub Actions（[build.yml](.github/workflows/build.yml)）が
[spinel](https://github.com/matz/spinel) をビルドしてバイナリを作り、
Releases に添付する。

```bash
git tag v1.0.1 && git push origin v1.0.1
```

対象は Linux x86-64（WSL 用）と macOS arm64 の 2 つ。タグを打たずに試すときは
Actions タブから `workflow_dispatch` で実行すると artifact ができる。

ワークフローにはスモークテストが入っていて、コンパイルしたバイナリに
実際にアクセスコードを流し込み、conf が書き換わることと既存エントリが
残ることを確認してから artifact を上げる。

## ローカルでビルドする

```bash
git clone --depth 1 https://github.com/matz/spinel.git
cd spinel && make deps && make      # bin/spinel ができる
cd ..
spinel/bin/spinel patch_access_code.rb -o patch_access_code
```

spinel はクロスコンパイルしないので、動かす環境そのものの上でビルドする。

CRuby でもそのまま実行できるので、動作確認だけなら spinel は要らない。

```bash
ruby patch_access_code.rb -n -f access_codes.txt path/to/BambuStudio.conf
```

## なぜ Windows ネイティブでなく WSL なのか

spinel は Windows ネイティブのバイナリを出力できない。runtime に `_WIN32` の
分岐が無く、生成される C が `sys/mman.h`・`pwd.h`・`fnmatch.h`・`sys/wait.h`
などを無条件で include するため、mingw ではコンパイルが通らない。これは
ビルド設定の問題ではなく runtime 移植が必要な話なので、spinel 公式の方針どおり
WSL 上で動かす。WSL からは Windows の C ドライブが `/mnt/c` として見えるので、
Windows 側の設定ファイルはそのまま書き換えられる。

## 実装メモ

spinel は Ruby の AOT コンパイラで、使えるのは型推論が通る範囲の Ruby に限られる。
この用途で引っかかった点:

**JSON ライブラリが無い。** conf をパースして書き戻すのではなく、
`user_access_code` のオブジェクトを波括弧の対応を数えて特定し、その範囲だけを
テキストとして差し替えている。結果として他の設定は 1 バイトも変わらない。
アクセスコード一覧の入力形式を空白区切りのプレーンテキストにしたのも同じ理由で、
これなら数行のトークン走査で読める。当初は JSON も受け付けていたが、
「引用符があれば JSON」という判定だとコメントに引用符が入っただけで
誤判定するため、形式は 1 つに絞った。

**`Dir.glob` は末尾要素のワイルドカードしか展開しない。** CRuby と違い
`/mnt/c/Users/*/AppData/...` がマッチしないので、`Dir.entries` で
ユーザーディレクトリを自前で走査している。

# Bambu Studio アクセスコード書き戻しツール

Bambu Studio が LAN モードのアクセスコードを保存し損ねる問題
([BambuStudio#5707](https://github.com/bambulab/BambuStudio/issues/5707)) の回避策。
`access_codes.json` のコードを `BambuStudio.conf` の `user_access_code` に書き戻す。

| ファイル | 用途 | Git |
|---|---|---|
| `patch_access_code.rb` | 書き戻し処理の本体。[spinel](https://github.com/matz/spinel) でバイナリ化する | 管理下 |
| `build_patcher.rb` | `access_codes.json` を埋め込んで spinel でビルドする | 管理下 |
| `launch_bambu_studio.rb` | macOS 用。コードを書き戻して Bambu Studio を起動する | 管理下 |
| `access_codes.json` | アクセスコードの一覧（唯一の情報源） | **除外** |
| `build/` | コードを埋め込んだビルド用ソース（自動生成） | **除外** |
| `dist/` | コードを埋め込んだバイナリ | **除外** |

バイナリはアクセスコードを埋め込み済みなので、単体で動く。Ruby も
`access_codes.json` も配布先には要らない。

アクセスコードを含むものは `.gitignore` で除外してある。クローンした直後は
`access_codes.json` が無いので、次の形式で自分で用意する。

```json
{
  "0309FA5XXXXXXXX": "1a2b3c4d",
  "01P09C4XXXXXXXX": "12345678"
}
```

シリアル番号とアクセスコードは Bambu Studio の
「デバイス」→ 対象プリンタ →「ネットワーク」で確認できるほか、
すでに一度接続済みなら `BambuStudio.conf` の `user_access_code` に入っている。

---

## Windows で使う（WSL 経由）

### なぜ WSL なのか

spinel は Windows ネイティブのバイナリを出力できない。runtime に `_WIN32` の
分岐が無く、生成される C が `sys/mman.h`・`pwd.h`・`fnmatch.h`・`sys/wait.h`
などを無条件で include するため、mingw ではコンパイルが通らない。これは
ビルド設定の問題ではなく runtime 移植が必要な話なので、spinel 公式の方針どおり
WSL 上で動かす。WSL からは Windows の C ドライブが `/mnt/c` として見えるので、
Windows 側の設定ファイルはそのまま書き換えられる。

### 前提

- WSL2 がインストール済み（PowerShell で `wsl -l -v` で確認。無ければ `wsl --install`）
- x86-64 の Windows（ARM 版 Windows では後述の手順で再ビルドが必要）

### 手順

**1. Bambu Studio を終了する**

これは必須。Bambu Studio は終了時に `BambuStudio.conf` を自分の内容で
上書きするので、起動したまま書き換えても消える。タスクトレイに常駐して
いないかも確認する。

**2. `dist/patch_access_code_wsl` を Windows 機に渡す**

USB メモリでも共有フォルダでも構わない。ここでは `ダウンロード` に
置いた前提で進める。

**3. WSL を開き、バイナリを WSL 側にコピーして実行する**

```bash
cp /mnt/c/Users/$USER/Downloads/patch_access_code_wsl ~/
chmod +x ~/patch_access_code_wsl
~/patch_access_code_wsl
```

`/mnt/c` から直接実行できることも多いが、ドライブのマウント設定によっては
実行権限を付けられないため、WSL 側のホームに置くのが確実。
Windows のユーザー名と WSL のユーザー名が違う場合は `$USER` を実際の
Windows ユーザー名に読み替える。

引数なしで実行すると `/mnt/c/Users/` 配下を走査して
`AppData/Roaming/BambuStudio/BambuStudio.conf` を自動で見つける。

```
対象: /mnt/c/Users/ユーザー名/AppData/Roaming/BambuStudio/BambuStudio.conf
  01P09C4XXXXXXXX => 12345678
  0309FA5XXXXXXXX => 1a2b3c4d
  0309FA5YYYYYYYY => 87654321
  （以下 access_codes.json の全件）
既存 2 件 / 追加 7 件 / 更新 0 件 → 合計 9 件
バックアップ: /mnt/c/Users/ユーザー名/AppData/Roaming/BambuStudio/BambuStudio.conf.bak-20260727083225
アクセスコードを書き込みました。
```

**4. Bambu Studio を起動して、プリンタに LAN モードで接続できることを確認する**

### オプション

書き込まずに結果だけ見る:

```bash
~/patch_access_code_wsl -n
```

conf のパスを明示する（自動検出が外れた場合や、別ユーザーの設定を触る場合）:

```bash
~/patch_access_code_wsl "/mnt/c/Users/別のユーザー/AppData/Roaming/BambuStudio/BambuStudio.conf"
```

### 元に戻す

実行のたびにタイムスタンプ付きのバックアップが同じディレクトリにできる。

```bash
ls /mnt/c/Users/$USER/AppData/Roaming/BambuStudio/BambuStudio.conf.bak-*
cp /mnt/c/Users/$USER/AppData/Roaming/BambuStudio/BambuStudio.conf.bak-20260727083225 \
   /mnt/c/Users/$USER/AppData/Roaming/BambuStudio/BambuStudio.conf
```

### うまくいかないとき

**`BambuStudio.conf が見つかりませんでした`**

Bambu Studio を一度も起動していないか、パスが標準と違う。PowerShell で
`echo $env:APPDATA` を実行して場所を確かめ、手順 3 のパス指定で渡す。

**`Permission denied`**

`/mnt/c` から直接実行しようとしている可能性が高い。手順 3 のとおり
WSL 側のホームにコピーしてから `chmod +x` する。

**`cannot execute binary file: Exec format error`**

ARM 版 Windows で x86-64 バイナリを動かそうとしている。下の再ビルド手順を
WSL 内でそのまま実行すれば、その環境向けのバイナリができる。

**書き換えたのに Bambu Studio でコードが消えている**

Bambu Studio が起動したままだった。終了させてから実行し直す。

---

## macOS で使う

```bash
./dist/patch_access_code       # 書き込む
./dist/patch_access_code -n    # 確認だけ
```

`~/Library/Application Support/BambuStudio/BambuStudio.conf` を自動検出する。
起動までまとめてやるなら従来どおり `ruby launch_bambu_studio.rb` でもよい。

---

## アクセスコードを追加・変更したら

`access_codes.json` を編集してから再ビルドする。`patch_access_code.rb` の
`CODES` は空のままで正しい。`build_patcher.rb` がコードを埋め込んだコピーを
`build/patch_access_code.gen.rb` に生成し、そちらをコンパイルする。
ソースを直接書き換えないので、リポジトリに認証情報が入らない。

バイナリはそれを動かす OS 上でビルドする必要がある（spinel はクロスコンパイル
しない）。macOS で作れば macOS 用、WSL で作れば WSL 用になる。

まず spinel を用意する。

```bash
git clone --depth 1 https://github.com/matz/spinel.git
cd spinel && make deps && make      # bin/spinel ができる
```

**macOS でビルドする**

```bash
SPINEL=/path/to/spinel/bin/spinel ruby build_patcher.rb
# → dist/patch_access_code
```

**WSL 内でビルドする**（ARM 版 Windows など、配布バイナリが動かない場合）

```bash
sudo apt install -y build-essential git ruby     # ruby は build_patcher.rb 用
SPINEL=/path/to/spinel/bin/spinel ruby build_patcher.rb -o dist/patch_access_code_wsl
```

`patch_access_code.rb` を spinel に直接渡してもコンパイルは通るが、
`CODES` が空なので何も書き込まないバイナリになる。必ず `build_patcher.rb`
を通すこと（Ruby が必要なのはこのビルド時だけで、できたバイナリの実行には要らない）。

**WSL 用バイナリを macOS から作る（Docker）**

```bash
docker run --rm --platform linux/amd64 -v "$PWD:/src" -w /src debian:bookworm-slim bash -c '
  apt-get update -qq && apt-get install -y -qq build-essential curl git ca-certificates
  git clone -q --depth 1 https://github.com/matz/spinel.git /build
  cd /build && make deps && make -j8
  /build/bin/spinel /src/patch_access_code.rb -o /src/dist/patch_access_code_wsl'
```

---

## 実装メモ

spinel は Ruby の AOT コンパイラで、使えるのは型推論が通る範囲の Ruby に限られる。
この用途で引っかかった点は 2 つ。

**JSON ライブラリが無い。** そのため conf をパースして書き戻すのではなく、
`user_access_code` のオブジェクトを波括弧の対応を数えて特定し、その範囲だけを
テキストとして差し替えている。結果として他の設定は 1 バイトも変わらない。
既存のエントリは読み取ってマージするので、`access_codes.json` に無いコードが
conf にあっても消えない。

**`Dir.glob` は末尾要素のワイルドカードしか展開しない。** CRuby と違い
`/mnt/c/Users/*/AppData/...` がマッチしないので、`Dir.entries` で
ユーザーディレクトリを自前で走査している。

---

## 取り扱い注意

`access_codes.json`、`build/`、`dist/` のバイナリはいずれもプリンタの
認証情報そのもの。これらは `.gitignore` で除外済みなので、通常の操作で
リポジトリに入ることはない。

ただし除外されるのはリポジトリの中だけで、ビルドしたバイナリを人に渡せば
アクセスコードを渡すのと同じになる。配布先には注意する。

`patch_access_code.rb` の `CODES` を手で埋めると、その瞬間から認証情報が
Git の管理対象になる。コードの追加・変更は必ず `access_codes.json` 側で行う。

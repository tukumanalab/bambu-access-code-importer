# 開発者向けメモ

使い方は [README.md](README.md) を参照。ここにはビルドの手順と
実装上の判断を書いておく。

## リポジトリ構成

| ファイル | 用途 |
|---|---|
| `main.go` | 登録処理の本体 |
| `console_windows.go` / `console_other.go` | コンソールのコードページと端末判定 |
| `main_test.go` | 一覧の読み取りと conf の書き換えのテスト |
| `build.sh` | アクセスコードを埋め込んだ配布用バイナリを作る |
| `.github/workflows/build.yml` | テストとクロスビルドの確認（配布物は作らない） |

## 配布物は手元で作る

配るバイナリは `./build.sh` で作り、ラボ内からしか見えない場所に置く。
リリースはしないので、GitHub の Releases は使わない。

CI はテストと、全ターゲットのクロスビルドが通ることの確認だけを行う。
埋め込みは `build.sh` でしかしないので、**CI の成果物にもリポジトリにも
認証情報は入らない**。手元は macOS なので、Windows 向けのコード
（`console_windows.go`）が壊れていないかは CI でしか気づけない。

```bash
./build.sh                    # access_codes.txt を埋め込む
./build.sh path/to/list.txt   # 一覧の場所を指定する
./build.sh --no-codes         # 埋め込まない版
```

対象は Windows x64 / Windows x86 / Windows arm64 / macOS arm64 / macOS x64 /
Linux x86-64。Go はクロスコンパイルするので、どれも 1 台の Mac から作れる。

配るのは Windows x86 と macOS の 2 つ（Intel Mac がいれば 3 つ）だけにする。
32bit の exe はどの Windows でも動くので、利用者に CPU を判別させずに済む。
似た名前が並ぶと取り違えが起きる（実際に arm64 版を実行して
「このアプリはお使いの PC で実行できません」になった）。

`build.sh` は全ターゲットを作る前に、ネイティブ版をその場でビルドして
`-n`（dry-run）で走らせ、埋め込んだ一覧が期待どおり読めることを確かめる。
書式が崩れていて 0 件になったバイナリを配ってしまう事故を防ぐため。

### 埋め込みの仕組み

一覧を base64 にして `-ldflags "-X main.embeddedCodesB64=..."` で渡している。
base64 にするのは、空白と改行が消えて `-X` にそのまま載せられるから。

作業ツリーには一切書き出さないので、ビルドの副産物として一覧がコミット
されることはない。ただし**埋め込んだコードは `strings` で取り出せる**。
難読化ではないので、置き場所で守るという前提を崩さないこと。

## conf のキーの意味（Bambu Studio 側の実装）

アクセスコードに関係するキーが 3 つある。挙動は BambuStudio のソース
（`src/slic3r/GUI/DeviceCore/DevManager.cpp`、`src/slic3r/GUI/DeviceManager.cpp`）で確認した。

| キー | 意味 |
|---|---|
| `user_access_code` | ユーザーが入力したアクセスコードの保管場所。**このツールが書くのはここ** |
| `access_code` | Studio がそのプリンタとの通信に使う作業用のコピー |
| `user_access_dev_ip` | プリンタの IP を `slicer_uuid` で暗号化したもの |

プリンタが LAN 上で見つかると、Studio は `DevManager.cpp` で
`access_code` と `user_access_code` の両方を dev_id 引きで読み込む。
したがって**同じネットワークにいて発見できるプリンタなら、
`user_access_code` に書いておくだけでコードが効く**。これが一括登録が
成立する根拠。

`user_access_dev_ip` は `restore_local_machines_from_user_access_config()`
専用で、発見を待たずに IP へ直接つなぎ直す近道に使われる。この関数は
`user_access_dev_ip` が無いエントリを `continue` で読み飛ばすため、
ここに値が無いプリンタはこの経路では復元されない（発見経由なら問題ない）。
値は `BBLCrossTalk::Decode_DevIp(encoded, slicer_uuid)` で復号される
インストール固有の暗号文なので、**このツールから生成することはできない**。

### 登録したコードが消える条件

`DevManager.cpp` に次の処理がある。

```cpp
if (obj && obj->is_cloud_mode_printer()) {
    obj->erase_user_access_code();
    obj->erase_user_access_dev_ip();
}
```

`erase_user_access_code()` は `app_config->erase("user_access_code", dev_id)`
を呼ぶので、**クラウドモードとして認識されたプリンタのエントリは
conf から削除される**。登録したのに件数が減っている場合は、
そのプリンタが LAN オンリーモードになっていない。

## 実装メモ

**conf は JSON としてパースし直さない。** `user_access_code` のオブジェクトを
波括弧の対応を数えて特定し、その範囲だけをテキストとして差し替えている。
`encoding/json` で読み書きするとキーの順序やエスケープの流儀が変わって
しまうが、この方法なら他の設定は 1 バイトも変わらない。値の中の波括弧を
数えないよう、文字列リテラルの中は読み飛ばす（テストあり）。

**説明文をエントリと誤読しない条件。** 一覧の行を採用するのは、トークン
（英数字・`_`・`-` の連続）がちょうど 2 つ、区切りが空白だけ、どちらも 6 文字
以上のときだけ。`# 以降`はコメント。区切りと長さを見ないと「3F 実習室 4台」
から `3F => 4` を拾ってしまい、埋め込み方式ではそれがそのまま conf に入る。

**一覧の取得元の優先順位。** `-f` → 埋め込み → 実行ファイルの隣またはカレント
ディレクトリの `access_codes.txt` → 実行時の貼り付け。実行ファイルの隣を見るのは、
Finder やエクスプローラから起動するとカレントディレクトリが別の場所になるため。

**Windows のコンソール。** 旧来の conhost は既定のコードページが CP932 で、
UTF-8 の日本語が化ける。起動時に `SetConsoleOutputCP(65001)` を呼んでいる。

**ダブルクリック起動。** 終了時に窓が閉じて結果が読めなくなるので、端末に
繋がっているときだけ Enter を待つ。判定は Windows では `GetConsoleMode` の
成否、それ以外ではキャラクタデバイスかどうか（`/dev/null` は `os.SameFile`
で除く）。パイプ経由の CI では待たない。

## 検討して見送った案

**spinel（Ruby AOT コンパイラ）。** 最初の実装はこれだった。Windows ネイティブ
のバイナリを出力できず（runtime に `_WIN32` の分岐が無く、生成される C が
`sys/mman.h`・`pwd.h` などを無条件で include する）、Windows のメンバーには
WSL の導入を求めることになっていた。これが敷居として重かったので Go に移した。
Go ならクロスコンパイルできるので、ARM 版 Windows・Intel Mac 用に各自が
ビルドする案内も不要になった。

**ブラウザ版（GitHub Pages + File System Access API）。** インストール不要で
最も敷居が低いが、**書き込めない**。Chromium はブロックリストで
`%APPDATA%`（`DIR_ROAMING_APP_DATA`）と macOS の `~/Library` を
`kBlockAllChildren` として弾く
（[`chrome_file_system_access_permission_context.cc`](https://chromium.googlesource.com/chromium/src/+/main/chrome/browser/file_system_access/chrome_file_system_access_permission_context.cc)）。
`BambuStudio.conf` はどちらの直下にもあるので、ファイル選択の時点で拒否される。
Safari・Firefox がそもそも `showOpenFilePicker` を持たない点も併せて見送り。

**一覧を実行時に URL から取得する。** 1 ファイル配布のまま、プリンタを増やした
ときに再配布が要らなくなる。埋め込みで運用してみて、更新のたびの配り直しが
負担になるようならこちらに移る。

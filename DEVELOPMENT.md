# 開発者向けメモ

使い方は [README.md](README.md) を参照。ここにはビルドの手順と
実装上の判断を書いておく。

## リポジトリ構成

| ファイル | 用途 |
|---|---|
| `main.go` | 登録処理の本体 |
| `console_windows.go` / `console_other.go` | コンソールのコードページと端末判定 |
| `main_test.go` | 一覧の読み取りと conf の書き換えのテスト |
| `build.sh` | 配布用のバイナリと、隣に添える一覧を `dist/` に用意する |
| `.github/workflows/build.yml` | テストとクロスビルドの確認（配布物は作らない） |

## 配布物は手元で作る

配るバイナリは `./build.sh` で作り、`access_codes.txt` と一緒にラボ内から
しか見えない場所に置く。バイナリだけは Releases にも置いてよい（後述）。
**`access_codes.txt` は Releases に絶対に載せない**。

CI はテストと、全ターゲットのクロスビルドが通ることの確認だけを行う。
**バイナリにアクセスコードは入らない**（実行時に隣のファイルを読む）ので、
CI の成果物にもリポジトリにも認証情報は入らない。手元は macOS なので、
Windows 向けのコード（`console_windows.go`）が壊れていないかは CI でしか
気づけない。

```bash
./build.sh                    # access_codes.txt を dist/ に添える
./build.sh path/to/list.txt   # 一覧の場所を指定する
./build.sh --no-codes         # バイナリだけ作る
```

対象は Windows x64 / Windows x86 / Windows arm64 / macOS arm64 / macOS x64 /
Linux x86-64。Go はクロスコンパイルするので、どれも 1 台の Mac から作れる。

配るのは Windows x86 と macOS の 2 つ（Intel Mac がいれば 3 つ）だけにする。
32bit の exe はどの Windows でも動くので、利用者に CPU を判別させずに済む。
似た名前が並ぶと取り違えが起きる（実際に arm64 版を実行して
「このアプリはお使いの PC で実行できません」になった）。

`build.sh` は全ターゲットを作る前に、ネイティブ版をその場でビルドして
`-n`（dry-run）で走らせ、配る一覧が期待どおり読めることを確かめる。
書式が崩れていて 0 件になった一覧を配ってしまう事故を防ぐため。

### 一覧をバイナリに埋め込まない理由

以前は `-ldflags "-X main.embeddedCodesB64=..."` で base64 にした一覧を
埋め込んでいたが、隣のファイルを読む方式に変えた。

- プリンタが増えたときに `access_codes.txt` を差し替えるだけで済み、
  6 ターゲット分のバイナリを作り直して配り直す必要がない。
- 埋め込んだコードはどのみち `strings` で取り出せたので、秘匿の強さは
  変わらない。どちらも**置き場所で守る**という前提は同じ。
- 一覧が平文で見えるぶん、配る前に中身を確認しやすい。

引き換えに、配布物が 2 ファイルになり、利用者が同じフォルダに置く手間が
増える。ダウンロードフォルダに両方入れれば済む程度なので、これを受け入れた。

## リリースする

バイナリが認証情報を持たなくなったので、Releases でも配れる。CI のタグ連動
リリースは外してあるので、手元から `gh` で作る。

```bash
./build.sh --no-codes                       # 一覧を添えずにバイナリだけ作る
git tag vX.Y.Z && git push origin vX.Y.Z    # main.go の const version と揃える
gh release create vX.Y.Z \
  dist/patch_access_code-windows-x86.exe \
  dist/patch_access_code-macos-arm64 \
  dist/patch_access_code-macos-x64
```

添付するのは利用者が選ぶ 3 つだけ。`--no-codes` を使うのは、`dist/` に
`access_codes.txt` を置かないため。**一覧は Releases に載せない**。載せたら
アクセスコードの公開になる。

タグは `main.go` の `const version` と一致させる。v1.0.x は spinel 版のタグで、
Go 版は 2.0.0 から。

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
から `3F => 4` を拾ってしまい、そのまま conf に入る。

**一覧の取得元の優先順位。** `-f` → 実行ファイルの隣の `access_codes.txt` →
カレントディレクトリの `access_codes.txt` → 実行時の貼り付け。実行ファイルの隣を
先に見るのは、Finder やエクスプローラから起動するとカレントディレクトリが別の
場所になるため（`os.Executable()` + `EvalSymlinks` で求める）。見つからなければ
探した場所を表示する。置き忘れが典型的な失敗なので、貼り付けを求める前に出す。

**conf の探し先を環境変数だけに頼らない。** `os.UserConfigDir()` は Windows では
`%APPDATA%` をそのまま読む。この 2 つがずれている PC が実際にあり
（`%USERPROFILE%` は `C:\Users\ラボメン`、`%APPDATA%` は `C:\Users\tukum\AppData\Roaming`）、
**Bambu Studio が読まないほうの conf を書き換えて「成功したのに反映されない」**
という状態になった。ホーム（`os.UserHomeDir()`）から組み立てた場所も候補に加え、
同じパスは 1 件に畳む。

**候補が複数あるときは選んでもらう。** 以前は黙って 1 件目を使っていたので、
取り違えに気づけなかった。更新日時の新しい順に並べて番号で選ばせる。Bambu Studio
は終了時に conf を書き直すので、実際に使われているものがいちばん新しい。パイプ
経由（CI のスモークテスト）では聞かずに 1 件目を使う。

**標準入力は 1 つの Reader で読む。** `bufio` の読み手を作り直すと、先読みして
バッファに残った分を取りこぼす。conf を選んだ直後に一覧を貼り付ける流れで実際に
壊れるので、パッケージ変数の `stdin` を全員で使う。

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
ときに一覧の配り直しも要らなくなる。ただしラボ内限定で取得できる置き場所と
その認証が要る。隣のファイルを読む方式で運用してみて、それでも配り直しが
負担になるようならこちらに移る。

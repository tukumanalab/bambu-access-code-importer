# Bambu Studio アクセスコード一括登録ツール

共用の 3D プリンタが何台もある環境で、**全台分の LAN モード用アクセスコードを
Bambu Studio にまとめて登録する**ツール。一度実行すれば、以後どのプリンタにも
コードを打ち込まずに LAN モードで接続できる。

**まだ一度も接続していないプリンタも先に登録できる**のが要点で、新しく入った
メンバーの PC でも、最初から全台が使える状態にしておける。

あわせて、Bambu Studio が LAN モードのアクセスコードを保存せず毎回入力を
求めてくる問題
([BambuStudio#5707](https://github.com/bambulab/BambuStudio/issues/5707) — 未解決)
の回避策にもなる。仕組みとしては `BambuStudio.conf` の `user_access_code` に
一覧を書き込んでいるだけで、他の設定には触れない。

配るのは**実行ファイルとアクセスコード一覧 `access_codes.txt` の 2 つ**。
同じフォルダに置いて実行すると、隣の一覧を読んで書き込む。コードの入力も
貼り付けも要らない。Windows はダブルクリックするだけ、Mac は OS の制限があり
初回だけターミナルでの操作が要る。

> **注意:** `access_codes.txt` はアクセスコードそのもの。ラボ内からしか
> アクセスできない場所に置くこと。このリポジトリにコードは入っていない。

## 使い方（使う人の手順）

登録した内容は Bambu Studio を再起動しても残るので、実行は基本的に一度でよい
（プリンタが増えて配布物が更新されたら、また実行する）。

### 1. Bambu Studio を終了する

これは必須。Bambu Studio は終了時に `BambuStudio.conf` を自分の内容で
上書きするので、起動したまま書き換えても消える。画面を閉じただけでは
終了していないことがあるので、タスクトレイ（時計の左の隠れたアイコン）や
Dock に残っていないかも確認する。

### 2. 配布場所から、自分の PC 用のファイルと `access_codes.txt` を落とす

| PC | ファイル |
|---|---|
| Windows | `bambu_access_code-windows-x64.exe` |
| Mac（M1 以降） | `bambu_access_code-macos-arm64` |
| Mac（Intel） | `bambu_access_code-macos-x64` |

**`access_codes.txt` も一緒にダウンロードし、実行ファイルと同じフォルダに
置く**（どちらもダウンロードフォルダに入れておけばよい）。実行ファイルは
自分の隣にあるこのファイルからアクセスコードを読む。

Windows 用は 64bit の exe を 1 つだけ配る。ARM の Windows でも動く（x64 を
エミュレーションで実行できるため）。32bit 版の `-windows-x86.exe` も作ってあるが、
**実行のたびに管理者のパスワードを聞かれる**ので、ふだんは使わない（理由は
「うまくいかないとき」を参照）。

`-windows-x86.exe` が要るのは、32bit しかない古い Windows と、x64 を
エミュレーションできない古い ARM 版 Windows だけ。**取り違えると「このアプリは
お使いの PC で実行できません」で止まる**ので、配布場所には 1 つだけ置くほうがよい。

### 3. 実行する

**Windows:** ダウンロードしたファイルをダブルクリックする。

「Windows によって PC が保護されました」と出たら、**「詳細情報」→「実行」**
と進む。作成元が登録されていないソフトすべてに出る警告で、初回だけ聞かれる。

**Mac:** ダウンロードしただけでは実行できない（実行の許可が付かず、
ダブルクリックしても開かないか、テキストとして開かれる）。Spotlight
（`⌘ + スペース`）で「ターミナル」を開き、次の 3 行を 1 行ずつ実行する。
M1 以降の Mac の場合。Intel Mac ならファイル名を `-macos-x64` に読み替える。

```bash
cd ~/Downloads
chmod +x bambu_access_code-macos-arm64        # 実行の許可を与える
xattr -c bambu_access_code-macos-arm64        # ダウンロード時に付く検疫マークを外す
./bambu_access_code-macos-arm64
```

`chmod` と `xattr` を済ませたファイルは、次回からダブルクリックでも起動できる。

先にダブルクリックしてしまうと「"bambu_access_code-macos-arm64" は開いていません」
という警告が出るが、**「ゴミ箱に入れる」は押さず「完了」で閉じて**、上の 3 行を
実行すればよい（[うまくいかないとき](#うまくいかないとき)）。

### 4. 表示を確認して Enter

設定ファイルを自動で見つけ、隣の `access_codes.txt` の一覧を書き込む。

```
bambu_access_code 2.0.1
対象: C:\Users\ユーザー名\AppData\Roaming\BambuStudio\BambuStudio.conf
一覧を使います: C:\Users\ユーザー名\Downloads\access_codes.txt
  01P09C4XXXXXXXX => 12345678
  0309FA5XXXXXXXX => 1a2b3c4d
既存 2 件 / 追加 7 件 / 更新 0 件 → 合計 9 件
バックアップ: ...\BambuStudio.conf.bak-20260729185633
アクセスコードを書き込みました。
Bambu Studio を起動して、LAN モードで接続できることを確認してください。

Enter キーを押すと終了します...
```

conf に元からあったエントリはマージされて消えない。

### 5. Bambu Studio を起動して、プリンタに LAN モードで接続できることを確認する

## 設定ファイル `BambuStudio.conf` の場所

書き込み先は自動で見つかるので、ふだんは気にしなくてよい。自動検出が外れて
パスを引数で渡すときと、バックアップから戻すときに要る。

| OS | 場所 |
|---|---|
| Windows | `%APPDATA%\BambuStudio\BambuStudio.conf`<br>実体は `C:\Users\ユーザー名\AppData\Roaming\BambuStudio\BambuStudio.conf` |
| Mac | `~/Library/Application Support/BambuStudio/BambuStudio.conf`<br>`~` は自分のホーム、つまり `/Users/ユーザー名` |

どちらも親フォルダが標準では隠れているので、そのまま辿ろうとすると見つからない。

**Windows:** `Windows キー + R` を押し、`%APPDATA%\BambuStudio` と入力して Enter。
そのフォルダが直接開く。エクスプローラで辿るなら、「表示」→「隠しファイル」に
チェックを入れないと `AppData` が出てこない。

**Mac:** Finder で `⌘ + ⇧ + G` を押し、
`~/Library/Application Support/BambuStudio` と入力して Enter。
メニューの「移動」に「ライブラリ」が出るのは Option キーを押している間だけ。

ファイル自体が無い場合は、Bambu Studio をまだ一度も起動していない。起動して
終了すると作られるので、そのうえで実行する。

## 配る人の手順

### 準備（一度だけ）

全台分のアクセスコードを、1 行に 1 台「シリアル番号 アクセスコード」を
空白区切りで並べた `access_codes.txt` にまとめる。

```
# 3F 実習室
0309FA5XXXXXXXX 1a2b3c4d
01P09C4XXXXXXXX 12345678

# 4F 工房 (# 以降はコメント、空行は無視される)
0309FA5YYYYYYYY 87654321
```

シリアル番号とアクセスコードは Bambu Studio の
「デバイス」→ 対象プリンタ →「ネットワーク」で確認できる。

この `access_codes.txt` は `.gitignore` で除外してあり、コミットされない。

### ビルドして配る

[Go](https://go.dev/dl/) を入れて（`brew install go`）、リポジトリの直下で
実行する。

```bash
./build.sh                    # access_codes.txt を dist/ に添える
./build.sh path/to/list.txt   # 一覧の場所を指定する
```

`dist/` に全 OS 分のバイナリと `access_codes.txt` のコピーができるので、
**ラボ内からしかアクセスできない場所**に置く。ビルドの途中で一覧の中身と
件数が表示されるので、配る前にそこを確認する。

置くのは使う人が選ぶ 3 つ（Windows は `-windows-x64.exe`、Mac は
`-macos-arm64` と `-macos-x64`）と `access_codes.txt` だけにする。似た名前の
ファイルが並んでいると取り違えが起きる。**32bit 版は置かない**（実行のたびに
管理者のパスワードを求められ、そのまま進むと管理者の設定ファイルが書き換わる）。

バイナリは認証情報を持たないので、
[Releases](https://github.com/tukumanalab/bambu-access-code-importer/releases)
からも落とせる。その場合も `access_codes.txt` はラボ内の置き場所から取る。
**一覧を Releases に載せてはいけない。**

プリンタを増やしたときは、`access_codes.txt` に 1 行足して置き場所の
`access_codes.txt` を差し替える。実行ファイルは一覧を持たないので、
**バイナリの作り直しも配り直しも要らない**。各自が新しい一覧をダウンロード
し直して実行すれば行き渡る。

## オプション

ふだんは要らないが、うまくいかないときの手掛かりになる。

```bash
bambu_access_code --version          # バージョンを表示する
bambu_access_code --help             # オプションの一覧を表示する
bambu_access_code -n                 # 書き込まずに結果だけ見る
bambu_access_code -d                 # どこをどう探したかを表示する
bambu_access_code -f 別の一覧.txt      # 隣の access_codes.txt 以外を使う
bambu_access_code <conf のパス>       # 対象を明示する（自動検出が外れた場合）
```

一覧は「`-f` で指定したファイル → 実行ファイルと同じフォルダの
`access_codes.txt` → カレントディレクトリの `access_codes.txt`」の順に探す。
どれも無い場合は、探した場所を表示したうえで、実行時に一覧の貼り付けを求める。

## 元に戻す

実行のたびにタイムスタンプ付きのバックアップが conf と同じディレクトリにできる。

```bash
cp BambuStudio.conf.bak-20260729185633 BambuStudio.conf
```

## うまくいかないとき

**`BambuStudio.conf が見つかりませんでした`**

Bambu Studio を一度も起動していないか、パスが標準と違う。探した場所が画面に
出るので、[設定ファイル `BambuStudio.conf` の場所](#設定ファイル-bambustudioconf-の場所)
と見比べて、実際の場所をパスとして引数で渡す。

**`access_codes.txt が見つかりませんでした`**

実行ファイルと `access_codes.txt` が別のフォルダにある。探した場所が画面に
出るので、そこに `access_codes.txt` を置いて実行し直す。Windows で
`access_codes.txt.txt` になっていないか（拡張子が隠れている場合に起きる）も
確認する。

**`"bambu_access_code-macos-arm64" は開いていません` と出る（Mac）**

「マルウェアが含まれていないことを検証できませんでした」と続く警告。ダウンロード
したファイルに付く**検疫属性（quarantine）**が残っているために出る。Apple の署名
（有料の開発者登録が要る）を付けていないソフトすべてに出るもので、ファイルが
壊れているわけではない。

**「ゴミ箱に入れる」は押さず、「完了」で閉じる。** そのうえでターミナルを開き、
検疫属性を外してから実行する。M1 以降の Mac の場合。Intel Mac ならファイル名を
`-macos-x64` に読み替える。

```bash
cd ~/Downloads
xattr -d com.apple.quarantine bambu_access_code-macos-arm64   # 検疫属性を外す
chmod +x bambu_access_code-macos-arm64                        # 実行の許可を与える
./bambu_access_code-macos-arm64
```

`xattr -c`（付いている属性を全部消す）でもよい。一度外せば以後この警告は出ず、
ダブルクリックでも起動できる。

ターミナルを使いたくない場合は、警告を「完了」で閉じた直後に**システム設定 →
プライバシーとセキュリティ**を開き、下のほうに出る「"bambu_access_code-macos-arm64"
は…ブロックされました」の右の**「このまま開く」**を押してもよい。ただしこの方法
でも、実行の許可（`chmod +x`）はターミナルで与える必要がある。

**実行しようとすると管理者のパスワードを聞かれる（Windows）**

`-windows-x86.exe`（32bit 版）を使っている。**`-windows-x64.exe` に替えれば出なく
なる。**

**そのまま管理者として実行してはいけない。** 昇格したプロセスは管理者アカウントの
ものとして動くため、いま使っている人ではなく**管理者の設定ファイルを書き換えて
しまう**（画面には成功と出るのに、Bambu Studio には反映されない）。実際にこれで
半日つまずいた。

Windows は、マニフェストを持たない 32bit の実行ファイルをインストーラかもしれない
ものとして扱い、昇格を求めることがある（ファイル名に `patch` / `install` /
`setup` / `update` を含む場合が典型だが、名前を変えても消えないことがある）。
64bit の実行ファイルはこの判定の対象外。

どうしても 32bit 版を使う必要があり、昇格を避けられない場合は、書き込み先を
明示すれば正しい conf を更新できる。

```powershell
.\bambu_access_code-windows-x86.exe "C:\Users\<自分のユーザー名>\AppData\Roaming\BambuStudio\BambuStudio.conf"
```

**`アクセスコードを書き込みました` と出たのに、Bambu Studio に反映されない**

書き換えた conf が、Bambu Studio の読む conf と別物の可能性がある。まず上の
「管理者のパスワードを聞かれる」に当てはまらないかを確認する。`対象:` の行に
`C:\Users\<自分のユーザー名>\...` 以外が出ていたら、それが起きている。

`-d` を付けると、どこをどう探したかが表示される。

```powershell
.\bambu_access_code-windows-x64.exe -d -n
```

候補が複数見つかったときは番号で聞くので、**更新日時がいちばん新しいもの**を選ぶ
（Bambu Studio は終了時に conf を書き直すので、使われているものが新しい）。
パスを直接指定してもよい。

```powershell
.\bambu_access_code-windows-x64.exe "$env:USERPROFILE\AppData\Roaming\BambuStudio\BambuStudio.conf"
```

**書き換えたのに Bambu Studio でコードが消えている**

まず Bambu Studio が起動したままでなかったか確認する。終了時に設定を
自分の内容で上書きするので、起動中に書き換えても消える。

それでも消える場合は、**そのプリンタが LAN オンリーモードになっていない**。
Bambu Studio はクラウドモードのプリンタを見つけると、そのプリンタの
アクセスコードを設定から削除する仕様になっている。プリンタ本体の画面で
LAN オンリーモードを有効にしてから、もう一度実行する。

**登録したのに、まだアクセスコードを聞かれる**

このツールが書き込むのは、プリンタが**ネットワーク上で見つかったときに
参照される**保管場所。プリンタの電源が入っていて同じネットワークにいれば
コードは自動で使われるが、見つからない状態では効かない。プリンタの電源と
ネットワーク接続を確認する。

**`このアプリはお使いの PC で実行できません`（Windows）**

exe と PC の CPU が合っていない。ふつうは `-windows-x64.exe` で動く（ARM の
Windows でもエミュレーションで動く）。確かめたい場合は `Windows キー + Pause`
で開く画面の「システムの種類」を見る（`ARM ベース` なら `-windows-arm64.exe` も
使える）。

`-windows-x64.exe` で止まるのは、32bit しかない古い Windows か、x64 を
エミュレーションできない古い ARM 版 Windows。その場合だけ `-windows-x86.exe`
を使う（管理者のパスワードを聞かれる点は上記のとおり）。

どれでも同じ表示が出る場合は、ダウンロードの途中でファイルが壊れている。
取得し直す。

**画面が一瞬で閉じてしまう**

処理が終わると `Enter キーを押すと終了します...` で待つので、通常は閉じない。
それより前に閉じる場合は、ウイルス対策ソフトに止められている可能性がある。

---

ビルドの手順と実装上の判断は [DEVELOPMENT.md](DEVELOPMENT.md) にまとめてある。

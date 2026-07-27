# Bambu Studio アクセスコード書き戻しツール

Bambu Studio が LAN モードのアクセスコードを保存し損ねる問題
([BambuStudio#5707](https://github.com/bambulab/BambuStudio/issues/5707)) の回避策。
プリンタのアクセスコード一覧を `BambuStudio.conf` の `user_access_code` に書き戻す。

アクセスコードは**実行時に貼り付けて渡す**方式なので、配布バイナリにも
このリポジトリにも認証情報は一切含まれない。バイナリは
[Releases](../../releases) からダウンロードするだけで誰でも使える。

## 使い方

アクセスコード一覧は、1 行に 1 台「シリアル番号 アクセスコード」を空白区切りで
並べたテキストとして、メンバーだけが見られる場所
（プライベートな Wiki・Notion・パスワードマネージャーなど）に掲載しておく。

```
# 3F プリンタ (# 以降はコメント)
0309FA5XXXXXXXX 1a2b3c4d
01P09C4XXXXXXXX 12345678
```

シリアル番号とアクセスコードは Bambu Studio の
「デバイス」→ 対象プリンタ →「ネットワーク」で確認できる。
（旧形式の JSON を貼り付けても読み取れる。）

> **注意:** この JSON はプリンタの認証情報そのもの。誰でも見られる場所には
> 置かないこと。

### Windows で使う（WSL 経由）

このツールは Windows 版のバイナリを作れないため、Windows に標準で用意されている
Linux 環境（WSL）から実行する。WSL からは Windows のドライブが見えるので、
Windows 側の設定ファイルをそのまま書き換えられる。

**0. WSL が使えることを確認する（初回だけ）**

スタートメニューで「PowerShell」を開き、次を実行する。

```powershell
wsl -l -v
```

Ubuntu などの名前が表示されれば準備できている。「WSL がインストールされて
いません」といったメッセージが出たら、次を実行して PC を再起動する
（途中で Linux 側のユーザー名とパスワードを決めるよう求められる。
Windows のものと違って構わない）。

```powershell
wsl --install
```

**1. Bambu Studio を終了する**

これは必須。Bambu Studio は終了時に `BambuStudio.conf` を自分の内容で
上書きするので、起動したまま書き換えても消える。画面を閉じただけでは
終了していないことがあるので、タスクトレイ（時計の左の隠れたアイコン）に
残っていないかも確認する。

**2. [Releases](../../releases) から `patch_access_code-linux-x86_64` をダウンロードする**

ブラウザで開いて、最新版の Assets からファイル名をクリックする。
そのまま「ダウンロード」フォルダに保存する。

**3. WSL を開く**

スタートメニューで「Ubuntu」（または `wsl --install` で入れた
ディストリビューション名）を選ぶ。黒いターミナル画面が開く。

**4. バイナリを WSL 側にコピーして実行する**

次の 3 行を 1 行ずつ実行する。`<Windowsのユーザー名>` は
`C:\Users\` の下にある自分のフォルダ名に置き換える。

```bash
cp /mnt/c/Users/<Windowsのユーザー名>/Downloads/patch_access_code-linux-x86_64 ~/
chmod +x ~/patch_access_code-linux-x86_64
~/patch_access_code-linux-x86_64
```

- `/mnt/c` は WSL から見た Windows の C ドライブ。つまり 1 行目は
  「ダウンロードしたファイルを Linux 側にコピーする」という意味。
- 2 行目でファイルに実行の許可を与え、3 行目で実行する。
- ユーザー名が分からなければ、`ls /mnt/c/Users` を実行すると一覧が出る。
- `/mnt/c` から直接実行できることも多いが、ドライブのマウント設定によっては
  実行権限を付けられないため、WSL 側のホーム（`~/`）に置くのが確実。

引数なしで実行すると `/mnt/c/Users/` 配下を走査して
`AppData/Roaming/BambuStudio/BambuStudio.conf` を自動で見つける。

**5. アクセスコード一覧をコピーして、ターミナルに貼り付ける**

掲載場所から一覧をまるごとコピーし、プロンプトに貼り付けて Enter を押す。

> **貼り付けは `Ctrl+V` では効かない。** ターミナル内で**右クリック**するか、
> `Ctrl+Shift+V` を押す。

```
対象: /mnt/c/Users/ユーザー名/AppData/Roaming/BambuStudio/BambuStudio.conf

アクセスコード一覧を貼り付けてください。
1 行に 1 台、「シリアル番号 アクセスコード」を空白区切りで:
  0309FA5XXXXXXXX 1a2b3c4d
  01P09C4XXXXXXXX 12345678
貼り付け後、空行 (Enter) で確定します。
  01P09C4XXXXXXXX => 12345678
  0309FA5XXXXXXXX => 1a2b3c4d
既存 2 件 / 追加 7 件 / 更新 0 件 → 合計 9 件
バックアップ: .../BambuStudio.conf.bak-20260727083225
アクセスコードを書き込みました。
```

conf に元からあったエントリはマージされて消えない。

**6. Bambu Studio を起動して、プリンタに LAN モードで接続できることを確認する**

### macOS で使う

Apple シリコン（M1 以降）の Mac 向け。手順の考え方は Windows と同じで、
Bambu Studio を終了 → 実行 → 一覧を貼り付け、の順。

**1. Bambu Studio を終了する**（メニューの「Bambu Studio」→「終了」）

**2. [Releases](../../releases) から `patch_access_code-macos-arm64` をダウンロードする**

**3.「ターミナル」を開いて実行する**

Spotlight（`⌘ + スペース`）で「ターミナル」と入力して開き、次を実行する。

```bash
chmod +x ~/Downloads/patch_access_code-macos-arm64
~/Downloads/patch_access_code-macos-arm64
```

`~/Library/Application Support/BambuStudio/BambuStudio.conf` を自動検出する。

「開発元を確認できないため開けません」と止められたら、次を実行してから
やり直す（ダウンロードしたファイルに付く検疫マークを外す操作）。

```bash
xattr -d com.apple.quarantine ~/Downloads/patch_access_code-macos-arm64
```

**4. アクセスコード一覧を貼り付けて Enter**（`⌘ + V` で貼り付けられる）

**5. Bambu Studio を起動して、LAN モードで接続できることを確認する**

### オプション

書き込まずに結果だけ見る:

```bash
./patch_access_code-... -n
```

貼り付けの代わりに一覧のファイルを渡す:

```bash
./patch_access_code-... -f access_codes.txt
```

カレントディレクトリに `access_codes.txt`（または旧形式の
`access_codes.json`）があれば、貼り付けを求めずにそれを自動で使う。

conf のパスを明示する（自動検出が外れた場合や、別ユーザーの設定を触る場合）:

```bash
./patch_access_code-... "/mnt/c/Users/別のユーザー/AppData/Roaming/BambuStudio/BambuStudio.conf"
```

### 元に戻す

実行のたびにタイムスタンプ付きのバックアップが conf と同じディレクトリにできる。

```bash
cp BambuStudio.conf.bak-20260727083225 BambuStudio.conf
```

### うまくいかないとき

**`BambuStudio.conf が見つかりませんでした`**

Bambu Studio を一度も起動していないか、パスが標準と違う。PowerShell で
`echo $env:APPDATA` を実行して場所を確かめ、パスを引数で渡す。

**`Permission denied`（WSL）**

`/mnt/c` から直接実行しようとしている可能性が高い。WSL 側のホームに
コピーしてから `chmod +x` する。

**`cannot execute binary file: Exec format error`**

ARM 版 Windows で x86-64 バイナリを動かそうとしている。
[配布バイナリが動かない環境で使う](#配布バイナリが動かない環境で使うarm-版-windowsintel-mac)
の手順を WSL 内でそのまま実行すれば、その環境向けのバイナリができる。

**書き換えたのに Bambu Studio でコードが消えている**

Bambu Studio が起動したままだった。終了させてから実行し直す。

**アクセスコードが読み取れませんでした**

貼り付けた内容から「シリアル番号 アクセスコード」の組が見つからなかった。
一覧をまるごとコピーしているか確認する。1 行に空白区切りでちょうど
2 項目ある行だけがエントリとして読み取られる（`#` 以降はコメント）。

---

## 配布バイナリが動かない環境で使う（ARM 版 Windows・Intel Mac）

Releases に置いてあるのは Linux x86-64 と macOS arm64 の 2 つだけ。
`Exec format error` が出るなど、どちらも動かない環境では自分でビルドする
（[spinel](https://github.com/matz/spinel) はクロスコンパイルしないので、
動かす環境そのものの上でビルドする）。

```bash
# WSL の場合の事前準備
sudo apt install -y build-essential git

git clone --depth 1 https://github.com/matz/spinel.git
cd spinel && make deps && make      # bin/spinel ができる
cd ..
git clone https://github.com/tukumanalab/bambu-access-code-keeper.git
spinel/bin/spinel bambu-access-code-keeper/patch_access_code.rb -o patch_access_code
```

できた `patch_access_code` は、ダウンロードしたバイナリと同じように使える。

---

ビルドやリリースの手順、実装上の判断は [DEVELOPMENT.md](DEVELOPMENT.md) にまとめてある。

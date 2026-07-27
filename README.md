# Bambu Studio アクセスコード書き戻しツール

Bambu Studio が LAN モードのアクセスコードを保存し損ねる問題
([BambuStudio#5707](https://github.com/bambulab/BambuStudio/issues/5707)) の回避策。
プリンタのアクセスコード一覧を `BambuStudio.conf` の `user_access_code` に書き戻す。

アクセスコードは**実行時に貼り付けて渡す**方式なので、配布バイナリにも
このリポジトリにも認証情報は一切含まれない。バイナリは
[Releases](../../releases) からダウンロードするだけで誰でも使える。

## 使い方

アクセスコード一覧は次の形式の JSON で、メンバーだけが見られる場所
（プライベートな Wiki・Notion・パスワードマネージャーなど）に掲載しておく。

```json
{
  "0309FA5XXXXXXXX": "1a2b3c4d",
  "01P09C4XXXXXXXX": "12345678"
}
```

シリアル番号とアクセスコードは Bambu Studio の
「デバイス」→ 対象プリンタ →「ネットワーク」で確認できる。

> **注意:** この JSON はプリンタの認証情報そのもの。誰でも見られる場所には
> 置かないこと。

### Windows で使う（WSL 経由）

**1. Bambu Studio を終了する**

これは必須。Bambu Studio は終了時に `BambuStudio.conf` を自分の内容で
上書きするので、起動したまま書き換えても消える。タスクトレイに常駐して
いないかも確認する。

**2. [Releases](../../releases) から `patch_access_code-linux-x86_64` をダウンロードする**

**3. WSL を開き、バイナリを WSL 側にコピーして実行する**

```bash
cp /mnt/c/Users/$USER/Downloads/patch_access_code-linux-x86_64 ~/
chmod +x ~/patch_access_code-linux-x86_64
~/patch_access_code-linux-x86_64
```

`/mnt/c` から直接実行できることも多いが、ドライブのマウント設定によっては
実行権限を付けられないため、WSL 側のホームに置くのが確実。
Windows のユーザー名と WSL のユーザー名が違う場合は `$USER` を実際の
Windows ユーザー名に読み替える。

引数なしで実行すると `/mnt/c/Users/` 配下を走査して
`AppData/Roaming/BambuStudio/BambuStudio.conf` を自動で見つける。

**4. アクセスコード一覧をコピーして、ターミナルに貼り付ける**

掲載場所から JSON をまるごとコピーし、プロンプトに貼り付けて Enter を押す。

```
対象: /mnt/c/Users/ユーザー名/AppData/Roaming/BambuStudio/BambuStudio.conf

アクセスコード一覧 (JSON) を貼り付けてください。
例:
  {
    "0309FA5XXXXXXXX": "1a2b3c4d"
  }
貼り付け後、空行 (Enter) で確定します。
  01P09C4XXXXXXXX => 12345678
  0309FA5XXXXXXXX => 1a2b3c4d
既存 2 件 / 追加 7 件 / 更新 0 件 → 合計 9 件
バックアップ: .../BambuStudio.conf.bak-20260727083225
アクセスコードを書き込みました。
```

conf に元からあったエントリはマージされて消えない。

**5. Bambu Studio を起動して、プリンタに LAN モードで接続できることを確認する**

### macOS で使う

[Releases](../../releases) から `patch_access_code-macos-arm64` をダウンロードし、
同じ手順で実行する（Bambu Studio 終了 → 実行 → 貼り付け）。

```bash
chmod +x ~/Downloads/patch_access_code-macos-arm64
~/Downloads/patch_access_code-macos-arm64
```

`~/Library/Application Support/BambuStudio/BambuStudio.conf` を自動検出する。
初回は Gatekeeper に止められることがある。その場合は
「システム設定」→「プライバシーとセキュリティ」で許可するか、
`xattr -d com.apple.quarantine <バイナリ>` を実行する。

### オプション

書き込まずに結果だけ見る:

```bash
./patch_access_code-... -n
```

貼り付けの代わりに JSON ファイルを渡す:

```bash
./patch_access_code-... -f access_codes.json
```

カレントディレクトリに `access_codes.json` があれば、貼り付けを求めずに
それを自動で使う。

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

ARM 版 Windows で x86-64 バイナリを動かそうとしている。下の手動ビルドを
WSL 内でそのまま実行すれば、その環境向けのバイナリができる。

**書き換えたのに Bambu Studio でコードが消えている**

Bambu Studio が起動したままだった。終了させてから実行し直す。

**アクセスコードが読み取れませんでした**

貼り付けた内容に `"シリアル": "コード"` の形の文字列が無かった。
JSON をまるごとコピーしているか確認する。コードブロックの記号や
前後の説明文が混ざっていても、`"..."` の組さえあれば読み取れる。

---

## リリース（ビルド）

バイナリは GitHub Actions（[build.yml](.github/workflows/build.yml)）が
[spinel](https://github.com/matz/spinel) でビルドする。認証情報を埋め込まないので
CI ビルド・公開配布ができる。

新しいバージョンを出すにはタグを push するだけ:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

Linux x86-64（WSL 用）と macOS arm64 のバイナリが Releases に添付される。
タグなしで試すときは Actions タブから `workflow_dispatch` で実行すると
artifact ができる。

### 手動ビルド（ARM 版 Windows・Intel Mac など）

spinel はクロスコンパイルしないので、動かす環境上でビルドする。

```bash
# WSL の場合の事前準備
sudo apt install -y build-essential git

git clone --depth 1 https://github.com/matz/spinel.git
cd spinel && make deps && make      # bin/spinel ができる
cd ..
git clone https://github.com/tukumanalab/bambu-access-code-keeper.git
spinel/bin/spinel bambu-access-code-keeper/patch_access_code.rb -o patch_access_code
```

### なぜ WSL なのか

spinel は Windows ネイティブのバイナリを出力できない。runtime に `_WIN32` の
分岐が無く、生成される C が `sys/mman.h`・`pwd.h`・`fnmatch.h`・`sys/wait.h`
などを無条件で include するため、mingw ではコンパイルが通らない。これは
ビルド設定の問題ではなく runtime 移植が必要な話なので、spinel 公式の方針どおり
WSL 上で動かす。WSL からは Windows の C ドライブが `/mnt/c` として見えるので、
Windows 側の設定ファイルはそのまま書き換えられる。

---

## リポジトリ構成

| ファイル | 用途 |
|---|---|
| `patch_access_code.rb` | 書き戻し処理の本体。spinel でバイナリ化する |
| `launch_bambu_studio.rb` | macOS 用の補助。手元の `access_codes.json` を書き戻して Bambu Studio を起動する（Ruby 必要） |
| `.github/workflows/build.yml` | タグ push で各 OS 向けバイナリをビルドして Releases に添付 |

`access_codes.json`（`-f` やカレントディレクトリ検出で使う場合）は
`.gitignore` で除外してあり、コミットされない。

## 実装メモ

spinel は Ruby の AOT コンパイラで、使えるのは型推論が通る範囲の Ruby に限られる。
この用途で引っかかった点:

**JSON ライブラリが無い。** conf をパースして書き戻すのではなく、
`user_access_code` のオブジェクトを波括弧の対応を数えて特定し、その範囲だけを
テキストとして差し替えている。結果として他の設定は 1 バイトも変わらない。
貼り付けられたアクセスコード一覧も同じ理屈で、引用符で囲まれた文字列を
順に拾って解釈する（だから前後にゴミが混ざっていても動く）。

**`Dir.glob` は末尾要素のワイルドカードしか展開しない。** CRuby と違い
`/mnt/c/Users/*/AppData/...` がマッチしないので、`Dir.entries` で
ユーザーディレクトリを自前で走査している。

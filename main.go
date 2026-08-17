// BambuStudio.conf の "user_access_code" にアクセスコードをまとめて登録する。
//
// conf 全体を JSON としてパースし直さず、"user_access_code" のオブジェクト部分
// だけをテキストとして差し替える。こうすると他の設定は 1 バイトも変わらない
// (キーの順序・空白・エスケープの流儀を Bambu Studio 側に合わせたままにできる)。
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const version = "2.0.1"

// 標準入力は 1 つの Reader で読む。読み手を作り直すと、先読みして
// バッファに残った分を取りこぼす (conf を選んだ直後に一覧を貼り付ける
// 流れで実際に起きる)。
var stdin = bufio.NewReaderSize(os.Stdin, 64*1024)

const (
	key    = "user_access_code"
	indent = "    "
)

// アクセスコード一覧のファイル名。実行ファイルと同じ場所、次にカレント
// ディレクトリの順で探す。
const codesFileName = "access_codes.txt"

// シリアル番号とアクセスコードに使われる文字。
func isTokenChar(c byte) bool {
	return c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' ||
		c == '_' || c == '-'
}

// text 中の "key" に対応する JSON オブジェクトの { と } の位置を返す。
// 文字列リテラル内の波括弧は数えない。見つからなければ (-1, -1)。
func objectSpan(text, k string) (int, int) {
	at := strings.Index(text, `"`+k+`"`)
	if at < 0 {
		return -1, -1
	}
	rel := strings.Index(text[at:], "{")
	if rel < 0 {
		return -1, -1
	}
	openAt := at + rel

	depth := 0
	inStr := false
	esc := false
	for i := openAt; i < len(text); i++ {
		c := text[i]
		switch {
		case inStr:
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
		case c == '"':
			inStr = true
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return openAt, i
			}
		}
	}
	return -1, -1
}

// 波括弧の中身から "..." を順に取り出す。偶数番目がキー、奇数番目が値。
func scanStrings(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		if s[i] != '"' {
			i++
			continue
		}
		var buf strings.Builder
		j := i + 1
		for j < len(s) {
			c := s[j]
			if c == '"' {
				break
			}
			if c == '\\' && j+1 < len(s) {
				buf.WriteByte(s[j+1])
				j += 2
				continue
			}
			buf.WriteByte(c)
			j++
		}
		out = append(out, buf.String())
		i = j + 1
	}
	return out
}

// エントリとして認める最短の長さ。実物はシリアル番号 15 文字・
// アクセスコード 8 文字なので、これを下回る組は説明文からの誤検出とみなす。
const minTokenLen = 6

// 1 行から「シリアル番号 アクセスコード」の組を取り出す。# 以降はコメント。
//
// 説明文を誤ってエントリと解釈しないよう、条件を絞ってある。トークン
// (英数字・_・- の連続) がちょうど 2 つで、区切りが空白だけで、どちらも
// minTokenLen 以上であること。この 3 つ目までを課さないと、たとえば
// 「3F 実習室 4台」から 3F => 4 を拾ってしまう。
func entryFromLine(line string) (string, string, bool) {
	var tokens []string
	var buf strings.Builder
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '#' {
			break
		}
		if isTokenChar(c) {
			buf.WriteByte(c)
			continue
		}
		if c != ' ' && c != '\t' && c != '\r' {
			return "", "", false // 空白以外が挟まる行は一覧ではなく説明文
		}
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	if len(tokens) != 2 {
		return "", "", false
	}
	if len(tokens[0]) < minTokenLen || len(tokens[1]) < minTokenLen {
		return "", "", false
	}
	return tokens[0], tokens[1], true
}

// 貼り付け・ファイルのテキストからアクセスコード一覧を読み取る。
// 1 行 1 台の「シリアル番号 アクセスコード」(空白区切り)。
// 2 項目ちょうどの行だけを拾うので、見出しや説明文が混ざっていても動く。
func parseCodes(text string) map[string]string {
	codes := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		if serial, code, ok := entryFromLine(line); ok {
			codes[serial] = code
		}
	}
	return codes
}

func existingEntries(text string) map[string]string {
	entries := map[string]string{}
	openAt, closeAt := objectSpan(text, key)
	if openAt < 0 {
		return entries
	}
	parts := scanStrings(text[openAt+1 : closeAt])
	for i := 0; i+1 < len(parts); i += 2 {
		entries[parts[i]] = parts[i+1]
	}
	return entries
}

func render(entries map[string]string) string {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("{\n")
	for i, k := range keys {
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		fmt.Fprintf(&b, "%s%s\"%s\": \"%s\"%s\n", indent, indent, k, entries[k], comma)
	}
	b.WriteString(indent + "}")
	return b.String()
}

func patch(text string, entries map[string]string) (string, error) {
	openAt, closeAt := objectSpan(text, key)
	if openAt >= 0 {
		return text[:openAt] + render(entries) + text[closeAt+1:], nil
	}
	// キーごと無い場合は先頭の { の直後に差し込む
	head := strings.Index(text, "{")
	if head < 0 {
		return "", fmt.Errorf("JSON として読めません")
	}
	return text[:head+1] +
		"\n" + indent + `"` + key + `": ` + render(entries) + "," +
		text[head+1:], nil
}

// ---- conf の探索 ----

// WSL から見た Windows 側のユーザープロファイル。
const wslUsersRoot = "/mnt/c/Users"

func wslPaths() []string {
	var list []string
	entries, err := os.ReadDir(wslUsersRoot)
	if err != nil {
		return list
	}
	for _, e := range entries {
		list = append(list, filepath.Join(wslUsersRoot, e.Name(),
			"AppData/Roaming/BambuStudio/BambuStudio.conf"))
	}
	return list
}

// ホームディレクトリから組み立てた設定ディレクトリ。
// os.UserConfigDir() は環境変数をそのまま読むので、%APPDATA% が
// %USERPROFILE% とずれている PC では別人の場所を指すことがある
// (実際に、%APPDATA% だけが別フォルダを指していて、Bambu Studio が
// 読まない conf を書き換えていた事例があった)。ホーム由来の場所も候補に
// 入れておき、食い違ったときは利用者に選んでもらう。
func homeConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Roaming")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	default:
		return filepath.Join(home, ".config")
	}
}

// conf の探索順。Windows ではまずシェルに聞いた Roaming AppData
// (Bambu Studio と同じ求め方)、次に OS 標準の設定ディレクトリ
// (Windows: %APPDATA%, macOS: ~/Library/Application Support, Linux: ~/.config)、
// 次にホーム由来の同じ場所、最後に WSL から見た Windows 側。
func candidatePaths() []string {
	var dirs []string
	if dir := roamingAppDataDir(); dir != "" {
		dirs = append(dirs, dir)
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		dirs = append(dirs, dir)
	}
	if dir := homeConfigDir(); dir != "" {
		dirs = append(dirs, dir)
	}

	var list []string
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			list = append(list, p)
		}
	}
	for _, dir := range dirs {
		add(filepath.Join(dir, "BambuStudio", "BambuStudio.conf"))
	}
	for _, p := range wslPaths() {
		add(p)
	}
	return list
}

// 候補が複数あるときの選択。Bambu Studio は終了時に conf を書き直すので、
// 実際に使われているものがいちばん新しい。新しい順に並べて既定にするが、
// 黙って選ぶと今回のような取り違えに気づけないので必ず確認を取る。
func chooseConf(found []string) string {
	if len(found) == 1 {
		return found[0]
	}

	sort.SliceStable(found, func(i, j int) bool {
		return confModTime(found[i]).After(confModTime(found[j]))
	})

	fmt.Printf("BambuStudio.conf が %d 件見つかりました。\n", len(found))
	for i, p := range found {
		fmt.Printf("  %d) %s  (更新 %s)\n",
			i+1, p, confModTime(p).Format("2006-01-02 15:04:05"))
	}
	fmt.Println("Bambu Studio が実際に使っているのは、ふつういちばん新しいものです。")

	// 貼り付けやパイプで動かしているときは聞かずに 1 件目を使う。
	if !stdinIsTerminal() {
		fmt.Println("1 件目を使います: " + found[0])
		return found[0]
	}

	return found[readChoice(stdin, len(found))]
}

// 1 から n までの番号を読む。Enter だけなら 1 件目 (0 を返す)。
func readChoice(r *bufio.Reader, n int) int {
	for {
		fmt.Printf("どれを使いますか (1-%d、Enter で 1): ", n)
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			// 入力が尽きたときも既定に倒す。答えられない状況で
			// 聞き続けても、窓が流れるだけで何も選べない。
			return 0
		}
		if i, convErr := strconv.Atoi(line); convErr == nil && i >= 1 && i <= n {
			return i - 1
		}
		if err != nil {
			return 0
		}
		fmt.Println("番号で答えてください。")
	}
}

func confModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// ---- アクセスコード一覧の取得 ----

// 実行ファイルの置かれているディレクトリ。macOS の Finder や Windows の
// エクスプローラから起動するとカレントディレクトリが別の場所になるため、
// 隣に置いた access_codes.txt を見つけるにはこちらを見る必要がある。
func exeDir() string {
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		p = resolved
	}
	return filepath.Dir(p)
}

func codesFileCandidates() []string {
	var list []string
	if dir := exeDir(); dir != "" {
		list = append(list, filepath.Join(dir, codesFileName))
	}
	if wd, err := os.Getwd(); err == nil && wd != exeDir() {
		list = append(list, filepath.Join(wd, codesFileName))
	}
	return list
}

func readPastedCodes() map[string]string {
	fmt.Println()
	fmt.Println("アクセスコード一覧を貼り付けてください。")
	fmt.Println("1 行に 1 台、「シリアル番号 アクセスコード」を空白区切りで:")
	fmt.Println("  0309FA5XXXXXXXX 1a2b3c4d")
	fmt.Println("  01P09C4XXXXXXXX 12345678")
	fmt.Println("貼り付け後、空行 (Enter) で確定します。")

	var buf strings.Builder
	for {
		line, err := stdin.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" && buf.Len() > 0 {
			break
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		if err != nil {
			break
		}
	}
	return parseCodes(buf.String())
}

// 一覧の取得元は次の順に決まる:
//  1. -f FILE で指定したファイル (管理者向けの明示指定)
//  2. 実行ファイルの隣、またはカレントディレクトリの access_codes.txt
//  3. どれも無ければ、その場で貼り付けてもらう
func loadCodes(codesFile string) (map[string]string, error) {
	if codesFile != "" {
		raw, err := os.ReadFile(codesFile)
		if err != nil {
			return nil, fmt.Errorf("一覧ファイルを読めません: %s", codesFile)
		}
		return parseCodes(string(raw)), nil
	}
	candidates := codesFileCandidates()
	for _, p := range candidates {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		fmt.Println("一覧を使います: " + p)
		return parseCodes(string(raw)), nil
	}

	// 実行ファイルの隣に置き忘れているのが典型なので、探した場所を見せる。
	fmt.Println()
	fmt.Println(codesFileName + " が見つかりませんでした。探した場所:")
	for _, p := range candidates {
		fmt.Println("  " + p)
	}
	fmt.Println("実行ファイルと同じフォルダに " + codesFileName + " を置いて、もう一度実行してください。")
	return readPastedCodes(), nil
}

// ---- main ----

const usage = `使い方: patch_access_code [オプション] [BambuStudio.conf のパス]

  -n, --dry-run     書き込まずに結果だけ表示する
  -f, --codes FILE  アクセスコード一覧をファイルから読む
  -d, --debug       どこをどう探したかを表示する
  -v, --version     バージョンを表示する
  -h, --help        このヘルプを表示する`

// 探し先が期待と違うときの手掛かり。どの取得元が何を返したかを並べる。
// 環境変数と実際の場所が食い違う PC があり、表示を見ないと切り分けられない。
func printDiagnostics() {
	fmt.Println()
	fmt.Println("--- 探索の内訳 ---")
	fmt.Printf("  OS/CPU: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	for _, name := range []string{"APPDATA", "USERPROFILE", "HOME", "XDG_CONFIG_HOME"} {
		if v := os.Getenv(name); v != "" {
			fmt.Printf("  %%%s%%: %s\n", name, v)
		}
	}
	if dir := roamingAppDataDir(); dir != "" {
		fmt.Println("  シェルに聞いた Roaming AppData: " + dir)
	}
	if dir, err := os.UserConfigDir(); err == nil {
		fmt.Println("  os.UserConfigDir(): " + dir)
	}
	if dir := homeConfigDir(); dir != "" {
		fmt.Println("  ホーム由来: " + dir)
	}
	fmt.Println("  実行ファイルの場所: " + exeDir())

	fmt.Println("  conf の候補:")
	for _, p := range candidatePaths() {
		fmt.Println("    " + p + describeFile(p))
	}
	fmt.Println("  一覧の候補:")
	for _, p := range codesFileCandidates() {
		fmt.Println("    " + p + describeFile(p))
	}
	fmt.Println("------------------")
}

func describeFile(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "  [無し]"
	}
	return "  [あり 更新 " + info.ModTime().Format("2006-01-02 15:04:05") + "]"
}

func main() {
	setupConsole()
	code := run()
	pauseIfInteractive()
	os.Exit(code)
}

func run() int {
	fmt.Println("patch_access_code " + version)

	dryRun := false
	debug := false
	confPath := ""
	codesFile := ""
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--version":
			return 0
		case "-h", "--help":
			fmt.Println()
			fmt.Println(usage)
			return 0
		case "-n", "--dry-run":
			dryRun = true
		case "-d", "--debug":
			debug = true
		case "-f", "--codes":
			i++
			if i >= len(args) {
				fmt.Println("-f の後に一覧ファイルのパスを指定してください。")
				return 1
			}
			codesFile = args[i]
		default:
			confPath = args[i]
		}
	}

	if debug {
		printDiagnostics()
	}

	if confPath == "" {
		var found []string
		for _, p := range candidatePaths() {
			if _, err := os.Stat(p); err == nil {
				found = append(found, p)
			}
		}
		if len(found) == 0 {
			fmt.Println("BambuStudio.conf が見つかりませんでした。")
			fmt.Println("探した場所:")
			for _, p := range candidatePaths() {
				fmt.Println("  " + p)
			}
			fmt.Println("パスを引数で指定してください: patch_access_code <BambuStudio.conf のパス>")
			return 1
		}
		confPath = chooseConf(found)
		if len(found) > 1 {
			fmt.Println()
		}
	}

	text, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Println("ファイルを読めません: " + confPath)
		return 1
	}
	fmt.Println("対象: " + confPath)

	if !strings.Contains(string(text), "{") {
		fmt.Println("JSON として読めません。中身を確認してください。")
		return 1
	}

	codes, err := loadCodes(codesFile)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}
	if len(codes) == 0 {
		fmt.Println("アクセスコードが読み取れませんでした。")
		fmt.Println("1 行に 1 台、「シリアル番号 アクセスコード」の形式で用意してください。")
		return 1
	}

	before := existingEntries(string(text))
	merged := map[string]string{}
	for k, v := range before {
		merged[k] = v
	}
	added, changed := 0, 0
	for k, v := range codes {
		old, ok := merged[k]
		switch {
		case !ok:
			added++
		case old != v:
			changed++
		}
		merged[k] = v
	}

	updated, err := patch(string(text), merged)
	if err != nil {
		fmt.Println("設定の書き換えに失敗しました。")
		return 1
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("  " + k + " => " + merged[k])
	}
	fmt.Printf("既存 %d 件 / 追加 %d 件 / 更新 %d 件 → 合計 %d 件\n",
		len(before), added, changed, len(merged))

	if dryRun {
		fmt.Println("--dry-run のため書き込みませんでした。")
		return 0
	}

	backup := confPath + ".bak-" + time.Now().Format("20060102150405")
	if err := os.WriteFile(backup, text, 0o600); err != nil {
		fmt.Println("バックアップを作れませんでした: " + backup)
		return 1
	}
	if err := os.WriteFile(confPath, []byte(updated), 0o600); err != nil {
		fmt.Println("書き込めませんでした: " + confPath)
		fmt.Println("元の内容は " + backup + " に残してあります。")
		return 1
	}

	// 書いたつもりで書けていないことがある (権限、別プロセスによる復元など)。
	// 必ず読み直して確認する。
	check, err := os.ReadFile(confPath)
	if err != nil || string(check) != updated {
		fmt.Println("書き込みが反映されていません。Bambu Studio を終了してから試してください。")
		fmt.Println("元の内容は " + backup + " に残してあります。")
		return 1
	}

	fmt.Println("バックアップ: " + backup)
	fmt.Println("アクセスコードを書き込みました。")
	fmt.Println("Bambu Studio を起動して、LAN モードで接続できることを確認してください。")
	return 0
}

// ダブルクリックで起動されたときに結果を読めないまま窓が閉じるのを防ぐ。
// パイプ経由 (CI のスモークテストなど) では待たない。
func pauseIfInteractive() {
	if !stdinIsTerminal() {
		return
	}
	fmt.Println()
	fmt.Print("Enter キーを押すと終了します...")
	stdin.ReadString('\n')
}

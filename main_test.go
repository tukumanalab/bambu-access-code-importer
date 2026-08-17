package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseCodes(t *testing.T) {
	text := "# 3F 実習室\n" +
		"0309FA5XXXXXXXX 1a2b3c4d\n" +
		"01P09C4XXXXXXXX 12345678  # 行末コメント\r\n" +
		"\n" +
		"0309FA5ZZZZZZZZ 11112222 33334444\n" + // 3 項目は無視
		"0309FA5WWWWWWWW\n" + // 単独トークンも無視
		"0309FA5YYYYYYYY 87654321"

	codes := parseCodes(text)
	want := map[string]string{
		"0309FA5XXXXXXXX": "1a2b3c4d",
		"01P09C4XXXXXXXX": "12345678",
		"0309FA5YYYYYYYY": "87654321",
	}
	if len(codes) != len(want) {
		t.Fatalf("件数が違う: got %d want %d (%v)", len(codes), len(want), codes)
	}
	for k, v := range want {
		if codes[k] != v {
			t.Errorf("%s: got %q want %q", k, codes[k], v)
		}
	}
}

// 一覧に混ざった説明文をエントリと誤認しないこと。
func TestParseCodesRejectsProse(t *testing.T) {
	for _, line := range []string{
		"日本語の 説明文",
		"3F 実習室 4台",                      // 空白区切りで 2 トークンに見えるが説明文
		"1 3",                            // 短すぎる組
		"プリンタ: 0309FA5XXXXXXXX 1a2b3c4d", // 記号が挟まる行
	} {
		if codes := parseCodes(line + "\n"); len(codes) != 0 {
			t.Errorf("拾ってはいけない行を拾った: %q => %v", line, codes)
		}
	}
}

// 隣のファイルを読む方式では、探す場所と順番そのものが仕様。
// 実行ファイルの隣を先に見ないと、Finder やエクスプローラから起動したときに
// 一覧を見つけられない。
func TestCodesFileCandidates(t *testing.T) {
	got := codesFileCandidates()
	if len(got) == 0 {
		t.Fatal("候補が空になった")
	}
	for _, p := range got {
		if !filepath.IsAbs(p) {
			t.Errorf("絶対パスでない: %q", p)
		}
		if filepath.Base(p) != codesFileName {
			t.Errorf("ファイル名が違う: %q", p)
		}
	}
	if dir := exeDir(); dir != "" && filepath.Dir(got[0]) != dir {
		t.Errorf("実行ファイルの隣を先に見ていない: %v", got)
	}
	seen := map[string]bool{}
	for _, p := range got {
		if seen[p] {
			t.Errorf("同じ場所を 2 回探している: %q", p)
		}
		seen[p] = true
	}
}

// %APPDATA% が %USERPROFILE% とずれている PC があり、環境変数だけを見ると
// Bambu Studio が読まない conf を書き換えてしまう。ホーム由来の場所も候補に
// 入れる。ふつうは両者が一致するので、そのときは 1 件に畳まれること。
func TestCandidatePathsHasNoDuplicates(t *testing.T) {
	got := candidatePaths()
	if len(got) == 0 {
		t.Fatal("候補が空になった")
	}
	seen := map[string]bool{}
	for _, p := range got {
		if seen[p] {
			t.Errorf("同じ場所を 2 回探している: %q", p)
		}
		seen[p] = true
		if filepath.Base(p) != "BambuStudio.conf" {
			t.Errorf("ファイル名が違う: %q", p)
		}
	}
}

// 候補が複数あるとき、Bambu Studio が実際に使っているのは終了時に書き直された
// いちばん新しいもの。端末でなければ聞かずにそれを選ぶ。
func TestChooseConfPrefersNewest(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.conf")
	recent := filepath.Join(dir, "recent.conf")
	for _, p := range []string{old, recent} {
		if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}

	if got := chooseConf([]string{old, recent}); got != recent {
		t.Errorf("新しいほうを選んでいない: got %q want %q", got, recent)
	}
	// 並び順に依存しないこと。
	if got := chooseConf([]string{recent, old}); got != recent {
		t.Errorf("並び順で結果が変わった: got %q want %q", got, recent)
	}
}

func TestReadChoice(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{"\n", 0},        // Enter だけなら 1 件目
		{"2\n", 1},       //
		{" 3 \n", 2},     // 前後の空白は無視
		{"", 0},          // 入力が尽きたら既定
		{"x\n9\n2\n", 1}, // 範囲外・数字でない入力はやり直し
		{"0\n1\n", 0},    // 0 は範囲外
	} {
		got := readChoice(bufio.NewReader(strings.NewReader(tc.in)), 3)
		if got != tc.want {
			t.Errorf("入力 %q: got %d want %d", tc.in, got, tc.want)
		}
	}
}

func TestChooseConfSingle(t *testing.T) {
	if got := chooseConf([]string{"/a/BambuStudio.conf"}); got != "/a/BambuStudio.conf" {
		t.Errorf("1 件のときはそのまま返すべき: %q", got)
	}
}

const sampleConf = `{
    "app": "BambuStudio",
    "note": "波括弧 { } を含む値",
    "user_access_code": {
        "0EXIST1": "11111111"
    },
    "region": "Asia"
}
`

func TestExistingEntries(t *testing.T) {
	got := existingEntries(sampleConf)
	if len(got) != 1 || got["0EXIST1"] != "11111111" {
		t.Fatalf("既存エントリを読めていない: %v", got)
	}
}

func TestExistingEntriesMissingKey(t *testing.T) {
	if got := existingEntries(`{"region": "Asia"}`); len(got) != 0 {
		t.Fatalf("キーが無いのに読めてしまった: %v", got)
	}
}

// 差し替えるのは user_access_code のオブジェクトだけで、
// 他の設定は 1 バイトも変わらないこと。
func TestPatchLeavesOtherSettingsUntouched(t *testing.T) {
	entries := map[string]string{"0EXIST1": "11111111", "0309FA5TEST": "1a2b3c4d"}
	got, err := patch(sampleConf, entries)
	if err != nil {
		t.Fatal(err)
	}

	for _, keep := range []string{
		`    "app": "BambuStudio",`,
		`    "note": "波括弧 { } を含む値",`,
		`    "region": "Asia"`,
	} {
		if !strings.Contains(got, keep) {
			t.Errorf("他の設定が失われた: %q\n---\n%s", keep, got)
		}
	}
	// キーは昇順に並ぶ。
	want := "    \"user_access_code\": {\n" +
		"        \"0309FA5TEST\": \"1a2b3c4d\",\n" +
		"        \"0EXIST1\": \"11111111\"\n" +
		"    },"
	if !strings.Contains(got, want) {
		t.Errorf("書き出しが想定と違う:\n%s", got)
	}
}

func TestPatchInsertsMissingKey(t *testing.T) {
	got, err := patch(`{
    "region": "Asia"
}
`, map[string]string{"0309FA5TEST": "1a2b3c4d"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"0309FA5TEST": "1a2b3c4d"`) {
		t.Errorf("挿入されていない:\n%s", got)
	}
	if !strings.Contains(got, `"region": "Asia"`) {
		t.Errorf("元の設定が失われた:\n%s", got)
	}
	// 書き戻したものを読み直せること。
	if again := existingEntries(got); again["0309FA5TEST"] != "1a2b3c4d" {
		t.Errorf("読み直せない: %v", again)
	}
}

func TestPatchWithoutBrace(t *testing.T) {
	if _, err := patch("これは JSON ではない", map[string]string{"A": "B"}); err == nil {
		t.Error("JSON でない入力がエラーにならなかった")
	}
}

// 値の中の波括弧を数えてしまうと範囲がずれる。
func TestObjectSpanIgnoresBracesInStrings(t *testing.T) {
	conf := `{"user_access_code": {"A": "}{"}, "after": "x"}`
	openAt, closeAt := objectSpan(conf, key)
	if got := conf[openAt : closeAt+1]; got != `{"A": "}{"}` {
		t.Errorf("範囲がずれている: %q", got)
	}
}

func TestRoundTripKeepsEscapes(t *testing.T) {
	// エスケープを含む値を読んでも壊れないこと (現実の conf には出ないが、
	// scanStrings の素朴な実装で壊れやすい箇所なので押さえておく)。
	parts := scanStrings(`"a\"b": "c\\d"`)
	if len(parts) != 2 || parts[0] != `a"b` || parts[1] != `c\d` {
		t.Errorf("エスケープの解釈が違う: %q", parts)
	}
}

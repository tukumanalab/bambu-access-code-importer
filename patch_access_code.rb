#!/usr/bin/env ruby
# frozen_string_literal: true

# BambuStudio.conf の "user_access_code" にアクセスコードをまとめて登録する。
# spinel (https://github.com/matz/spinel) で単一バイナリにコンパイルして使う。
#
#   spinel patch_access_code.rb -o patch_access_code
#
# spinel には JSON ライブラリが無いため、conf 全体をパースし直さず
# "user_access_code" のオブジェクト部分だけをテキストとして差し替える。
# こうすると他の設定は 1 バイトも変わらない。

# アクセスコードは実行時に受け取る (優先順):
#   1. -f FILE で指定したファイル
#   2. カレントディレクトリの access_codes.txt
#   3. どちらも無ければ、その場で貼り付けてもらう
# 形式は 1 行 1 台の「シリアル番号 アクセスコード」(空白区切り)。
# バイナリにもこのソースにも認証情報は含まれないので、そのまま配布できる。

KEY = "user_access_code"
INDENT = "    "
TOKEN_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

# WSL から見た Windows 側のユーザープロファイルを走査する。
# spinel の Dir.glob は末尾要素のワイルドカードしか展開しないため、
# "/mnt/c/Users/*/AppData/..." は使えない。自前で列挙する。
def wsl_paths
  list = []
  root = "/mnt/c/Users"
  return list unless Dir.exist?(root)
  Dir.entries(root).each do |name|
    if name != "." && name != ".."
      p = root + "/" + name + "/AppData/Roaming/BambuStudio/BambuStudio.conf"
      list << p if File.exist?(p)
    end
  end
  list
end

# conf の探索順。Windows ネイティブ (APPDATA) → WSL (/mnt/c) → macOS。
def candidate_paths
  list = []
  appdata = ENV["APPDATA"]
  if appdata && appdata != ""
    list << appdata.tr("\\", "/") + "/BambuStudio/BambuStudio.conf"
  end
  wsl_paths.each do |p|
    list << p
  end
  home = ENV["HOME"]
  if home && home != ""
    list << home + "/Library/Application Support/BambuStudio/BambuStudio.conf"
  end
  list
end

# text 中の "key" に対応する JSON オブジェクトの { と } の位置を返す。
# 文字列リテラル内の波括弧は数えない。見つからなければ [-1, -1]。
def object_span(text, key)
  at = text.index("\"" + key + "\"")
  return [-1, -1] if at.nil?

  open_at = text.index("{", at)
  return [-1, -1] if open_at.nil?

  depth = 0
  in_str = false
  esc = false
  i = open_at
  n = text.length
  while i < n
    c = text[i]
    if in_str
      if esc
        esc = false
      elsif c == "\\"
        esc = true
      elsif c == "\""
        in_str = false
      end
    elsif c == "\""
      in_str = true
    elsif c == "{"
      depth += 1
    elsif c == "}"
      depth -= 1
      return [open_at, i] if depth == 0
    end
    i += 1
  end
  [-1, -1]
end

# 波括弧の中身から "..." を順に取り出す。偶数番目がキー、奇数番目が値。
def scan_strings(s)
  out = []
  i = 0
  n = s.length
  while i < n
    if s[i] == "\""
      buf = ""
      j = i + 1
      while j < n
        c = s[j]
        break if c == "\""
        if c == "\\"
          buf = buf + s[j + 1].to_s
          j += 2
        else
          buf = buf + c.to_s
          j += 1
        end
      end
      out << buf
      i = j + 1
    else
      i += 1
    end
  end
  out
end

def valid_token?(s)
  return false if s.length == 0
  i = 0
  while i < s.length
    return false if TOKEN_CHARS.index(s[i]).nil?
    i += 1
  end
  true
end

# 行からトークン (英数字・_・- の連続) を取り出す。# 以降はコメント。
def line_tokens(line)
  tokens = []
  buf = ""
  i = 0
  while i < line.length
    c = line[i]
    break if c == "#"
    if TOKEN_CHARS.index(c).nil?
      if buf != ""
        tokens << buf
        buf = ""
      end
    else
      buf = buf + c.to_s
    end
    i += 1
  end
  tokens << buf if buf != ""
  tokens
end

# 貼り付け・ファイルのテキストからアクセスコード一覧を読み取る。
# 1 行 1 台の「シリアル番号 アクセスコード」(空白区切り)。
# 2 項目ちょうどの行だけを拾うので、見出しや説明文が混ざっていても動く。
def parse_codes(text)
  map = {}
  start = 0
  n = text.length
  while start <= n
    nl = text.index("\n", start)
    line = nl.nil? ? text[start..-1].to_s : text[start...nl].to_s
    tokens = line_tokens(line)
    if tokens.length == 2
      k = tokens[0].to_s
      v = tokens[1].to_s
      map[k] = v if valid_token?(k) && valid_token?(v)
    end
    break if nl.nil?
    start = nl + 1
  end
  map
end

def read_pasted_codes
  puts ""
  puts "アクセスコード一覧を貼り付けてください。"
  puts "1 行に 1 台、「シリアル番号 アクセスコード」を空白区切りで:"
  puts "  0309FA5XXXXXXXX 1a2b3c4d"
  puts "  01P09C4XXXXXXXX 12345678"
  puts "貼り付け後、空行 (Enter) で確定します。"
  buf = ""
  depth = 0
  opened = false
  while (line = STDIN.gets)
    blank = line == "\n" || line == "\r\n" || line == ""
    break if blank && buf != ""
    buf = buf + line
    i = 0
    while i < line.length
      c = line[i]
      if c == "{"
        depth += 1
        opened = true
      elsif c == "}"
        depth -= 1
      end
      i += 1
    end
    break if opened && depth <= 0
  end
  parse_codes(buf)
end

def existing_entries(text)
  map = {}
  span = object_span(text, KEY)
  return map if span[0] < 0

  body = text[(span[0] + 1)...span[1]].to_s
  parts = scan_strings(body)
  i = 0
  while i + 1 < parts.length
    map[parts[i].to_s] = parts[i + 1].to_s
    i += 2
  end
  map
end

def render(map)
  keys = map.keys.sort
  body = ""
  i = 0
  keys.each do |k|
    comma = i == keys.length - 1 ? "" : ","
    body = body + INDENT + INDENT + "\"" + k + "\": \"" + map[k].to_s + "\"" + comma + "\n"
    i += 1
  end
  "{\n" + body + INDENT + "}"
end

def patch(text, map)
  span = object_span(text, KEY)
  if span[0] < 0
    # キーごと無い場合は先頭の { の直後に差し込む
    head = text.index("{")
    return "" if head.nil?
    return text[0..head].to_s +
           "\n" + INDENT + "\"" + KEY + "\": " + render(map) + "," +
           text[(head + 1)..-1].to_s
  end
  text[0...span[0]].to_s + render(map) + text[(span[1] + 1)..-1].to_s
end

# ---- main ----

dry_run = false
path = ""
codes_file = ""
i = 0
while i < ARGV.length
  a = ARGV[i]
  if a == "-n" || a == "--dry-run"
    dry_run = true
  elsif a == "-f" || a == "--codes"
    i += 1
    codes_file = ARGV[i].to_s
  else
    path = a
  end
  i += 1
end

if path == ""
  candidate_paths.each do |p|
    path = p if path == "" && File.exist?(p)
  end
end

if path == ""
  puts "BambuStudio.conf が見つかりませんでした。"
  puts "パスを引数で指定してください: patch_access_code <BambuStudio.conf のパス>"
  exit 1
end

unless File.exist?(path)
  puts "ファイルがありません: " + path
  exit 1
end

puts "対象: " + path

codes = {}
if codes_file != ""
  unless File.exist?(codes_file)
    puts "ファイルがありません: " + codes_file
    exit 1
  end
  codes = parse_codes(File.read(codes_file))
elsif File.exist?("access_codes.txt")
  puts "カレントディレクトリの access_codes.txt を使います。"
  codes = parse_codes(File.read("access_codes.txt"))
else
  codes = read_pasted_codes
end

if codes.empty?
  puts "アクセスコードが読み取れませんでした。"
  puts "1 行に 1 台、「シリアル番号 アクセスコード」の形式で貼り付けてください。"
  exit 1
end

text = File.read(path)
if text.index("{").nil?
  puts "JSON として読めません。中身を確認してください。"
  exit 1
end

before = existing_entries(text)
merged = {}
before.each do |k, v|
  merged[k] = v
end
added = 0
changed = 0
codes.each do |k, v|
  old = merged[k]
  if old.nil?
    added += 1
  elsif old != v
    changed += 1
  end
  merged[k] = v
end

updated = patch(text, merged)
if updated == ""
  puts "設定の書き換えに失敗しました。"
  exit 1
end

merged.keys.sort.each do |k|
  puts "  " + k + " => " + merged[k].to_s
end
puts "既存 " + before.length.to_s + " 件 / 追加 " + added.to_s + " 件 / 更新 " + changed.to_s + " 件 → 合計 " + merged.length.to_s + " 件"

if dry_run
  puts "--dry-run のため書き込みませんでした。"
  exit 0
end

backup = path + ".bak-" + Time.now.strftime("%Y%m%d%H%M%S")
File.write(backup, text)
File.write(path, updated)
puts "バックアップ: " + backup
puts "アクセスコードを書き込みました。"

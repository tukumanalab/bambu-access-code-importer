#!/usr/bin/env ruby
# frozen_string_literal: true

# access_codes.json のアクセスコードを patch_access_code.rb に埋め込み、
# spinel でバイナリにコンパイルする。
#
#   ruby build_patcher.rb                            # spinel は PATH から探す
#   SPINEL=/path/to/spinel/bin/spinel ruby build_patcher.rb
#   ruby build_patcher.rb -o dist/patch_access_code_wsl
#
# spinel には JSON パーサが無いため、コードはコンパイル時に埋め込むしかない。
# ただし patch_access_code.rb 自体は書き換えず、埋め込み済みのコピーを
# build/ に生成してそれをコンパイルする。こうすると認証情報が
# リポジトリ管理下のソースに入らない。

require "json"
require "fileutils"

DIR = __dir__
SRC = File.join(DIR, "patch_access_code.rb")
CODES_JSON = File.join(DIR, "access_codes.json")
GEN = File.join(DIR, "build", "patch_access_code.gen.rb")

out = "dist/patch_access_code"
spinel = ENV["SPINEL"] || "spinel"
args = ARGV.dup
while (i = args.index("-o"))
  out = args[i + 1] or abort "-o の後に出力先が必要です"
  args.delete_at(i + 1)
  args.delete_at(i)
end
spinel = args.shift unless args.empty?

abort "#{CODES_JSON} がありません" unless File.exist?(CODES_JSON)
codes = JSON.parse(File.read(CODES_JSON))
abort "access_codes.json が空です" if codes.empty?

# 生成コードに埋め込むので、Ruby の文字列リテラルを壊す文字は弾く
codes.each do |serial, code|
  unless serial.match?(/\A[A-Za-z0-9_-]+\z/) && code.match?(/\A[A-Za-z0-9_-]+\z/)
    abort "使えない文字が含まれています: #{serial.inspect} => #{code.inspect}"
  end
end

body = codes.map { |k, v| %(  "#{k}" => "#{v}") }.join(",\n")
block = "CODES = {\n#{body}\n}\n"

src = File.read(SRC)
marker = /(^# --- BEGIN GENERATED CODES ---\n).*?(^# --- END GENERATED CODES ---$)/m
abort "patch_access_code.rb に生成ブロックの目印が見つかりません" unless src.match?(marker)
generated = src.sub(marker) { "#{Regexp.last_match(1)}#{block}#{Regexp.last_match(2)}" }

FileUtils.mkdir_p(File.dirname(GEN))
File.write(GEN, generated)
File.chmod(0o600, GEN)

target = File.expand_path(out, DIR)
FileUtils.mkdir_p(File.dirname(target))

puts "#{codes.size} 件のアクセスコードを埋め込みました (#{GEN})"
unless system(spinel, GEN, "-o", target)
  abort "spinel でのコンパイルに失敗しました (spinel のパスを SPINEL= で指定してください)"
end
puts "生成: #{target}"

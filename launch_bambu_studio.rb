#!/usr/bin/env ruby
# frozen_string_literal: true

# Bambu Studio 起動前に、access_codes.json のアクセスコードを
# BambuStudio.conf の user_access_code へ注入してから Studio を起動する。
# Studio がコードの保存を落とすバグ (bambulab/BambuStudio#5707) の回避策。

require "json"
require "fileutils"

CONF = File.expand_path("~/Library/Application Support/BambuStudio/BambuStudio.conf")
CODES = File.expand_path("access_codes.json", __dir__)

# 起動中の Studio があれば終了を待つ(終了時に conf を上書きするため)
if system("pgrep -q BambuStudio")
  puts "Bambu Studio を終了しています..."
  system(%(osascript -e 'quit app "BambuStudio"'))
  30.times do
    break unless system("pgrep -q BambuStudio")
    sleep 1
  end
  abort "Bambu Studio が終了できませんでした。手動で終了してから再実行してください。" if system("pgrep -q BambuStudio")
end

codes = JSON.parse(File.read(CODES))
conf = JSON.parse(File.read(CONF))

FileUtils.cp(CONF, "#{CONF}.bak")

conf["user_access_code"] = (conf["user_access_code"] || {}).merge(codes)
File.write(CONF, JSON.generate(conf))

codes.each { |serial, code| puts "#{serial} => #{code}" }
puts "アクセスコードを注入しました。Bambu Studio を起動します。"
system("open", "-a", "BambuStudio")

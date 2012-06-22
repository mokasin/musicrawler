#!/usr/bin/env ruby

require 'listen'

def rebuild_haml(path)
	puts ">>> Change detected to: #{path}"
	newpath = path.chomp(File.extname(path)) + ".html"
	IO.popen("haml #{path} #{newpath}") do |io|
    print(io.readpartial(512)) until io.eof?
  end
end

Listen.to('html/templates', :filter => /\.haml$/) do |modified, added, removed|
	modified.each {|p| rebuild_haml(p)}
	added.each {|p| rebuild_haml(p)}
end

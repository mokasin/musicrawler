#!/usr/bin/env ruby

require 'listen'

def rebuild_haml(path)
	puts ">>> Change detected to: #{path}"
	newpath = "web/templates/" + path.chomp(File.extname(path)) + ".html"
	IO.popen("haml #{path} #{newpath}") do |io|
    print(io.readpartial(512)) until io.eof?
  end
end

def rebuild_less(path)
	puts ">>> Change detected to: #{path}"
	newpath = "web/assets/css/" + path.chomp(File.extname(path)) + ".css"
	IO.popen("lessc #{path} #{newpath}") do |io|
    print(io.readpartial(512)) until io.eof?
  end
end

Listen.to('webdev/templates', :filter => /\.haml$/) do |modified, added, removed|
	modified.each {|p| rebuild_haml(p)}
	added.each {|p| rebuild_haml(p)}
end

Listen.to('webdev/css/less', :filter => /\.less$/) do |modified, added, removed|
	modified.each {|p| rebuild_less(p)}
	added.each {|p| rebuild_less(p)}
end

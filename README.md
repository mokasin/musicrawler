musicrawler
===========

Description
-----------
musicrawler is a fast mp3/ogg/... indexer that offers it's service over the net
via HTTP/Json.

Currently tested in Linux. But there is no reason, other platform should not
work.

Status
------
Basic web access works.

Dependencies
------------
* [TagLib](http://developer.kde.org/~wheeler/taglib.html)
via the C-interface for reading tag metadata
* [gotaglib](http://github.com/mokasin/gotaglib)
* [go-sqlite3](https://github.com/mattn/go-sqlite3) by Yasuhiro Matsumoto
* [HAML](http://haml.info/)
* [LESS](http://lesscss.org/)
  (for [Bootstrap](http://twitter.github.com/bootstrap/))

Get it
------
1. Install *taglib* and *sqlite3* libraries.
2. If you haven't already, prepend a directory of your choice to GOPATH
   environment variable (see go help gopath for help) and run

		$ go get github.com/mokasin/musicrawler

	Get *HAML* and *LESS* via RubyGems and Nodejs Package Manager

		$ gem install haml
		$ npm -g install less

	or do it your own way. **haml** and **lessc** should be in an executable
	path.

Build
-----
Fetch go dependencies

	$ go get

and build it (on Linux) with

	$ ./make.sh

This compiles also less- and haml-files.

To build with debug symbols just

	$ go build

it yourself.

License
-------
GNU General Public License Version 3 or above
http://www.gnu.org/licenses/gpl.txt

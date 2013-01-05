musicrawler
===========

Description
-----------
musicrawler is a fast mp3/ogg/... indexer that offers its service over the net
via HTTP/Json.

Currently tested in Linux. However, there is no reason other platforms should not
work.

Status
------
Basic web access works.

Dependencies
------------
* [TagLib](http://taglib.github.com/)
via the C-interface for reading tag metadata
* [gotaglib](http://github.com/mokasin/gotaglib)
* [go-sqlite3](https://github.com/mattn/go-sqlite3) by Yasuhiro Matsumoto
* [gorilla/mux](https://code.google.com/p/gorilla/)
* [LESS](http://lesscss.org/)
  (for [Bootstrap](http://twitter.github.com/bootstrap/))

Get it
------
1. Install *taglib* and *sqlite3* libraries.
2. If you haven't already, prepend a directory of your choice to GOPATH
   environment variable (see go help gopath for help) and run

		$ go get github.com/mokasin/musicrawler

	Get *LESS* via Node.js Package Manager

		$ npm -g install less

	or do it your own way. **lessc** should be in an executable
	path.

Build
-----
Fetch go dependencies

	$ go get

get external libs as git submodules

	$ git submodule init
	$ git submodule update

and build it (on Linux) with

	$ ./make.sh

This compiles also less-files.

To build with debug symbols just

	$ go install

it yourself.

License
-------
GNU General Public License Version 3 or above
http://www.gnu.org/licenses/gpl.txt

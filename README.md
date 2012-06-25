musicrawler
===========

Description
-----------
musicrawler is a fast mp3/ogg/... indexer that offers it's service over the net
via HTTP/Json.

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

Build
-----

	$ ./make.sh

To build with debug symbols just

	$ go build

it yourself.

License
-------
GNU General Public License Version 3 or above
http://www.gnu.org/licenses/gpl.txt

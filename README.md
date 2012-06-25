musicrawler
===========

Description
-----------
musicrawler is fast mp3/ogg/… indexer that offers it's service over the net via
HTTP/Json.

Status
------

Basic web access works.

Dependencies
------------
* TagLib via the C-interface for reading tag metadata 
  (http://developer.kde.org/~wheeler/taglib.html)
* gotaglib
* go-sqlite3 by Yasuhiro Matsumoto (https://github.com/mattn/go-sqlite3)
* HAML
* LESS (for Bootstrap)

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

package index

const sql_insert_artist = "INSERT OR IGNORE INTO Artist(name) VALUES (?);"
const sql_insert_album = `INSERT OR IGNORE INTO Album(name, artist_id)
	VALUES (?,
			(SELECT ID FROM Artist WHERE name = ?));`

const sql_add_track = `INSERT INTO Track(
	path,
	title,
	album_id,
	tracknumber,
	year,
	length,
	genre,
	filemtime,
	dbmtime)
    VALUES( ?, ?, 
		   (SELECT ID FROM Album  WHERE name = ?), 
		    ?, ?, ?, ?, ?, ?);`

const sql_update_timestamp = "UPDATE Track SET dbmtime = ? WHERE path = ?;"

const sql_update_track = `UPDATE Track SET
	title       = ?,
	album_id    = (SELECT ID FROM Album  WHERE name = ?),
	tracknumber = ?,
	year        = ?,
	length		= ?,
	genre       = ?,
	filemtime   = ?
	WHERE path  = ?;`

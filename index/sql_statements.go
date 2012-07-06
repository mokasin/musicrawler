package index

// SQL queries to create the database schema
const sql_create_artist = `
	CREATE TABLE Artist
	(
		ID   INTEGER  NOT NULL PRIMARY KEY,
		name TEXT     UNIQUE
	);`
const sql_create_album = `
	CREATE TABLE Album
	(
		ID   INTEGER NOT NULL PRIMARY KEY,
		name TEXT,
		artist INTEGER REFERENCES Artist(ID) ON DELETE SET NULL
	);`
const sql_create_album_index = `
	CREATE UNIQUE INDEX 'album_artist' ON Album (name, artist);`
const sql_create_track = `
	CREATE TABLE Track
	(
		ID INTEGER NOT NULL PRIMARY KEY,
		path        TEXT NOT NULL,
		title       TEXT,
		tracknumber INTEGER,
		year        INTEGER,
		length      INTEGER,
		genre       TEXT,
		trackalbum	INTEGER	REFERENCES Album(ID) ON DELETE SET NULL,
		filemtime	INTEGER,
		dbmtime		INTEGER
	);`

const sql_insert_artist = "INSERT OR IGNORE INTO Artist(name) VALUES (?);"
const sql_insert_album = `INSERT OR IGNORE INTO Album(name, artist)
		VALUES (?,
				(SELECT ID FROM Artist WHERE name = ?));`

const sql_add_track = `INSERT INTO Track(
	path,
	title,
	trackalbum,
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
	trackalbum  = (SELECT ID FROM Album  WHERE name = ?),
	tracknumber = ?,
	year        = ?,
	length		= ?,
	genre       = ?,
	filemtime   = ?
	WHERE path  = ?;`

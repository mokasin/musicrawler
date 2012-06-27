package index

// SQL queries to create the database schema
const SQL_CREATE_ARTIST = `
	CREATE TABLE Artist
	(
		ID   INTEGER  NOT NULL PRIMARY KEY,
		name TEXT     UNIQUE
	);`
const SQL_CREATE_ALBUM = `
	CREATE TABLE Album
	(
		ID   INTEGER NOT NULL PRIMARY KEY,
		name TEXT    UNIQUE
	);`
const SQL_CREATE_TRACK = `
	CREATE TABLE Track
	(
		path        TEXT NOT NULL PRIMARY KEY,
		title       TEXT,
		tracknumber INTEGER,
		year        INTEGER,
		length      INTEGER,
		genre       TEXT,
		trackartist	INTEGER	REFERENCES Album(ID) ON DELETE SET NULL,
		trackalbum	INTEGER	REFERENCES Artist(ID) ON DELETE SET NULL,
		filemtime	INTEGER,
		dbmtime		INTEGER
	);`

const SQL_INSERT_ARTIST = "INSERT OR IGNORE INTO Artist(name) VALUES (?);"
const SQL_INSERT_ALBUM = "INSERT OR IGNORE INTO Album(name) VALUES (?);"

const SQL_ADD_TRACK = `INSERT INTO Track(
	path,
	title,
	trackartist,
	trackalbum,
	tracknumber,
	year,
	length,
	genre,
	filemtime,
	dbmtime)
    VALUES( ?, ?, 
		   (SELECT ID FROM Artist WHERE name = ?), 
		   (SELECT ID FROM Album  WHERE name = ?), 
		    ?, ?, ?, ?, ?, ?);`

const SQL_UPDATE_TIMESTAMP = "UPDATE Track SET dbmtime = ? WHERE path = ?;"

const SQL_UPDATE_TRACK = `UPDATE Track SET
	title       = ?,
	trackartist = (SELECT ID FROM Artist WHERE name = ?),
	trackalbum  = (SELECT ID FROM Album  WHERE name = ?),
	tracknumber = ?,
	year        = ?,
	length		= ?,
	genre       = ?,
	filemtime   = ?
	WHERE path  = ?;`

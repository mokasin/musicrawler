package main

import "testing"
import (
	"database/sql"
	"os"
)

const dbName = "test.db"

func TestCreateDatabase(t *testing.T) {
	if _, err := os.Stat(dbName); err == nil {
		t.Errorf("File %v does already exist. Please delete it.", dbName)
		return
	}

	i, err := NewIndex(dbName)
	defer os.Remove(dbName)

	if err != nil {
		t.Errorf("Database error: %v", err)
	}

	if i.Filename != dbName {
		t.Errorf("Filename not set correctly. %v vs. %v", i.Filename, dbName)
	}

	var path, name string
	var id int
	if err := i.db.QueryRow("SELECT path FROM Track").Scan(&path); err != sql.ErrNoRows {
		t.Errorf("%v", err)
	}
	if err := i.db.QueryRow("SELECT ID, name FROM Artist").Scan(&id, &name); err != sql.ErrNoRows {
		t.Errorf("%v", err)
	}
	if err := i.db.QueryRow("SELECT ID, name FROM Album").Scan(&id, &name); err != sql.ErrNoRows {
		t.Errorf("%v", err)
	}
}

func TestUpdatetrack(t *testing.T) {
	const mp3File = "test/test.mp3"

	i, _ := NewIndex(dbName)
	defer os.Remove(dbName)

	ti := &FileInfo{filename: mp3File, mtime: 123456}

	tx, err := i.db.Begin()
	if err != nil {
		t.Errorf("%v", err)
	}

	if err := i.addTrack(ti, tx); err != nil {
		t.Errorf("Adding track failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("%v", err)
	}

	var path string
	if err := i.db.QueryRow("SELECT path FROM Track").Scan(&path); err != nil {
		t.Errorf("No track in database: %v", err)
	}

	tag, err := i.GetTrackByPath(mp3File)

	var isexp = []struct {
		is  string
		exp string
	}{
		{tag.Artist, "TestArtist"},
		{tag.Album, "TestAlbum"},
		{tag.Title, "TestTitle"},
	}

	if err != nil {
		t.Errorf("Can't read database %v", err)
	}

	for _, tt := range isexp {
		if tt.is != tt.exp {
			t.Errorf("Is: %v, want: %v", tt.is, tt.exp)
		}
	}
}

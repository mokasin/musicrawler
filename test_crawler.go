package main

import (
	"math/rand"
	"musicrawler/source"
)

type testCrawler struct {
}

func randomString(length int) string {
	glyphs := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pw := make([]byte, length)
	for i := 0; i < length; i++ {
		pw[i] = glyphs[rand.Int()%len(glyphs)]
	}
	return string(pw)
}

const ARTISTNUMBER = 40
const ALBUMNUMBER = 80
const TRACKNUMBER = 1000

var artists = make([]string, ARTISTNUMBER)
var albums = make([]string, ALBUMNUMBER)

func init() {
	for i := 0; i < len(artists); i++ {
		artists[i] = randomString(5 + rand.Int()%26)
	}

	for i := 0; i < len(albums); i++ {
		albums[i] = randomString(5 + rand.Int()%26)
	}
}

func getArtist() string {
	return artists[rand.Int()%len(artists)]
}

func getAlbum() string {
	return albums[rand.Int()%len(albums)]
}

type TestInfo struct {
	path  string
	mtime int64
}

// Getter of TestInfo.filename
func (ti *TestInfo) Path() string {
	return ti.path
}

// Getter of TestInfo.mtime
func (ti *TestInfo) Mtime() int64 {
	return ti.mtime
}

// Reads tags (id3, vorbis,…) from file
func (ti *TestInfo) Tags() (*source.TrackTags, error) {

	return &source.TrackTags{
		Path:    ti.path,
		Title:   randomString(40),
		Artist:  getArtist(),
		Album:   getAlbum(),
		Comment: "",
		Genre:   "",
		Year:    2012,
		Track:   1,
		Bitrate: 128,
		Length:  400,
	}, nil
}

func (t *testCrawler) Crawl(tracks chan<- source.TrackInfo, done chan<- bool) {
	for i := int64(0); i < TRACKNUMBER; i++ {
		tracks <- &TestInfo{path: randomString(50 + rand.Int()%70), mtime: i}
	}
	done <- true
}

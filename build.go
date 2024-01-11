package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/h2non/filetype"
)

const (
	UNKNOWN int = iota
	MUSIC
	IMAGE
)

const (
	indexJsonFile     = "index.json"
	albumJsonFileName = "album.json"
	serverUrlTemplate = "https://cdn.jsdelivr.net/gh/yingshulu/content"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	build()
}

type Market struct {
	Albums []Cover `json:"albums"`
}

type Album struct {
	Cover
	Songs []Song `json:"songs"`
}

type Cover struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	ImageUrl    string `json:"image_url"`
	PlaylistUrl string `json:"playlist_url"`
}

type Song struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func build() {
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Panicf("build read current failure: %v", err)
	}

	market := &Market{}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			album, err := browserAlbumFile(e.Name())
			if err != nil {
				log.Printf("browser directory %s error %v", e.Name(), err)
				continue
			}

			album.Id = len(market.Albums)
			market.Albums = append(market.Albums, album.Cover)
			err = dumpAlbum(album)
			if err != nil {
				log.Printf("browser dump album %s error %v", e.Name(), err)
			}
		}
	}

	err = dumpStruct(market, indexJsonFile)
	if err != nil {
		log.Panic(err)
	}
}

func browserAlbumFile(dirName string) (*Album, error) {
	albumFileName := filePath(albumJsonFileName, dirName)
	album, err := parseAlbum(albumFileName)
	if err != nil {
		log.Printf("parseAlbum from json %s failed %v\n", albumFileName, err)
		album = &Album{}
		album.Name = dirName
		err = nil
	}

	isSongRecorded := func(name string) bool {
		for _, song := range album.Songs {
			if song.Name == name {
				return true
			}
		}
		return false
	}

	entries, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() || isSongRecorded(e.Name()) {
			continue
		}

		fileName := filePath(e.Name(), dirName)
		t := fileType(fileName)
		switch t {
		case MUSIC:
			index := len(album.Songs)
			album.Songs = append(album.Songs, newSong(index, e.Name(), dirName))
		case IMAGE:
			album.ImageUrl = fileUrl(e.Name(), dirName)
		default:
		}
	}
	album.PlaylistUrl = fileUrl(albumJsonFileName, dirName)
	return album, nil
}

func dumpAlbum(album *Album) error {
	name := filePath(albumJsonFileName, album.Name)
	return dumpStruct(album, name)
}

func dumpStruct(obj interface{}, fileName string) error {
	data, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, data, 0666)
	return err
}

func parseAlbum(name string) (*Album, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var album = &Album{}
	err = json.Unmarshal(content, album)
	return album, err
}

func newSong(id int, name, parent string) Song {
	return Song{
		Id:   id,
		Name: name,
		Url:  fileUrl(name, parent),
	}
}

func filePath(name, parent string) string {
	return fmt.Sprintf("%s/%s", parent, name)
}

func fileUrl(name, parent string) string {
	name, parent = url.PathEscape(name), url.PathEscape(parent)
	return fmt.Sprintf("%s/%s/%s", serverUrlTemplate, parent, name)
}

func fileType(name string) (typ int) {
	header := fileHeader(name)
	if header == nil {
		return
	}
	if filetype.IsImage(header) {
		typ = IMAGE
	}
	if filetype.IsAudio(header) {
		typ = MUSIC
	}
	return
}

func fileHeader(name string) []byte {
	f, err := os.Open(name)
	if err != nil {
		log.Printf("open file %s error %v\n", name, err)
		return nil
	}
	defer f.Close()

	header := make([]byte, 261)
	_, err = f.Read(header)
	if err != nil {
		log.Printf("read file %s error %s\n", name, err)
		return nil
	}
	return header
}

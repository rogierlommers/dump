package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
	"github.com/sirupsen/logrus"
)

var feed *feeds.Feed

func init() {
	logrus.Infof("setting up feed")

	feed = &feeds.Feed{
		Title:       "dump rss feed",
		Link:        &feeds.Link{Href: "http://dump.lommers.org"},
		Description: "exposes all shared files",
		Author:      &feeds.Author{Name: "Rogier", Email: ""},
		Created:     time.Now(),
	}

}

// type downloadedFile struct {
// 	UID     string `json:"uid"`
// 	Name    string `json:"name"`
// 	Referer string `json:"referer"`
// 	Size    int64  `json:"size"`
// }

func AddDownload(filename string, referer string, fullURL string) {
	newDownload := &feeds.Item{
		Title:       filename,
		Link:        &feeds.Link{Href: fullURL},
		Description: fmt.Sprintf("file %q has been downloaded by referer %q", filename, referer),
		Created:     time.Now(),
	}

	feed.Items = append(feed.Items, newDownload)
}

func RSSHandler(w http.ResponseWriter, req *http.Request) {
	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rss))
}

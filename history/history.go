package history

import (
	"net/http"

	"github.com/rogierlommers/tinycache"
	"github.com/sirupsen/logrus"
)

var history *tinycache.Cache

type download struct {
	filename string
	referer  string
	fullURL  string
}

func init() {
	logrus.Info("setting up cache")
	history = tinycache.NewCache(10)
}

func AddElement(filename string, referer string, fullURL string) {
	e := download{
		filename: filename,
		referer:  referer,
		fullURL:  fullURL,
	}

	logrus.Infof("download finished %v", e.filename)
	history.Add(e)
}

func HistoryHandler(w http.ResponseWriter, req *http.Request) {
	for _, d := range history.GetElements() {
		logrus.Infof("adding download: %v", d.(download))
	}

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(history))
}

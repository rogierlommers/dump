package history

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rogierlommers/tinycache"
	"github.com/sirupsen/logrus"
)

var history *tinycache.Cache

type download struct {
	Name              string    `json:"name"`
	Referer           string    `json:"referer"`
	RemoteAddress     string    `json:"remote_address"`
	TimestampDownload time.Time `json:"timestamp_download"`
}

func init() {
	logrus.Info("setting up cache")
	history = tinycache.NewCache(1000)
}

func AddElement(filename string, referer string, IP string) {
	e := download{
		Name:              filename,
		Referer:           referer,
		RemoteAddress:     IP,
		TimestampDownload: time.Now(),
	}

	logrus.Infof("download finished %v", e.Name)
	history.Add(e)
}

func HistoryHandler(w http.ResponseWriter, req *http.Request) {

	h, err := json.Marshal(history.GetElements())
	if err != nil {
		logrus.Errorf("history-listing error: %v", err)
		http.Error(w, "error exposing history", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h))
}

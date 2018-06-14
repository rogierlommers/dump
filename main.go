package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	host = "0.0.0.0"
	port = 8080
)

var uploadDir = "uploads"

// Request parameters
const (
	paramUuid = "qquuid" // uuid
	paramFile = "qqfile" // file name
)

// Chunked request parameters
const (
	paramPartIndex       = "qqpartindex"      // part index
	paramPartBytesOffset = "qqpartbyteoffset" // part byte offset
	paramTotalFileSize   = "qqtotalfilesize"  // total file size
	paramTotalParts      = "qqtotalparts"     // total parts
	paramFileName        = "qqfilename"       // file name for chunked requests
	paramChunkSize       = "qqchunksize"      // size of the chunks
)

type UploadResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	PreventRetry bool   `json:"preventRetry"`
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/chunksdone", ChunksDoneHandler)
	router.HandleFunc("/list", ListFilesHandler)
	router.HandleFunc("/download/{uid}", DownloadHandler)
	router.HandleFunc("/rss", RSSHandler)
	router.Handle("/upload/", http.StripPrefix("/upload/", http.HandlerFunc(UploadHandler)))
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static"))))

	logrus.Infof("deamon running on host %s and port %d", host, port)
	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}

}

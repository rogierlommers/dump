package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var uploadDir string

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
	debug := flag.Bool("debug", false, "set to true if you want debug info")
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	uploadDir = os.Getenv("UPLOADDIR")

	if err := checkDatadir(uploadDir); err != nil {
		logrus.Fatalf("invalid data directory: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/chunksdone", ChunksDoneHandler)
	router.HandleFunc("/list", ListFilesHandler)
	router.HandleFunc("/download/{uid}", DownloadHandler)
	router.HandleFunc("/rss", RSSHandler)
	router.Handle("/upload/", http.StripPrefix("/upload/", http.HandlerFunc(UploadHandler)))
	router.PathPrefix("/").Handler(http.StripPrefix("/landing-page", http.FileServer(http.Dir("static"))))

	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	logrus.Infof("using data directory: %s", uploadDir)
	logrus.Info("deamon running on host 0.0.0.0 and port 8080")

	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}

}

func checkDatadir(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("no env var UPLOADDIR defined")
	}

	if _, err := os.Stat(path); err != nil {
		return err
	}

	return nil
}

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	uploadDir string
	username  string
	password  string
)

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
	username = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")

	if err := checkDatadir(); err != nil {
		logrus.Fatalf("invalid data directory: %v", err)
	}

	if err := checkAuth(); err != nil {
		logrus.Fatalf("invalid credentials: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/chunksdone", ChunksDoneHandler)
	router.HandleFunc("/list", ListFilesHandler)
	router.HandleFunc("/rss", RSSHandler)
	router.HandleFunc("/download/{uid}", DownloadHandler)
	router.Handle("/upload/", http.StripPrefix("/upload/", http.HandlerFunc(UploadHandler)))
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static"))))

	router.Use(basicAuthMiddleware, loggingMiddleware)

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

func checkDatadir() error {
	if len(uploadDir) == 0 {
		return fmt.Errorf("no env var UPLOADDIR defined")
	}

	if _, err := os.Stat(uploadDir); err != nil {
		return err
	}

	return nil
}

func checkAuth() error {
	if len(username) == 0 || len(password) == 0 {
		return fmt.Errorf("no env var USERNAME or PASSWORD defined")
	}

	return nil
}

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.HasPrefix(r.RequestURI, "/download/") {

			logrus.Infof(r.RequestURI)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				http.Error(w, "Not authorized", 401)
				return
			}

			b, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				http.Error(w, err.Error(), 401)
				return
			}

			pair := strings.SplitN(string(b), ":", 2)
			if len(pair) != 2 {
				http.Error(w, "Not authorized", 401)
				return
			}

			if pair[0] != username || pair[1] != password {
				http.Error(w, "Not authorized", 401)
				return
			}
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("incoming request: %s", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

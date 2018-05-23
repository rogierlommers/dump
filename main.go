package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func DownloadHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uid := vars["uid"]

	download, err := strconv.ParseBool(req.URL.Query().Get("download"))
	if err != nil {
		logrus.Error(err)
	}

	logrus.Debugf("force download? %v", download)

	targetDir := filepath.Join(uploadDir, uid)
	logrus.Infof("download uid %q, directory %s", uid, targetDir)

	var filepathToDownload string
	var filenameToDownload string

	err = filepath.Walk(targetDir, func(listedPath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			logrus.Debugf("about to download: %q", listedPath)
			filepathToDownload = listedPath
			filenameToDownload = info.Name()
			return nil
		}

		return nil
	})

	if err != nil {
		http.Error(w, "error reading file from disk", http.StatusBadRequest)
		return
	}

	// now download file
	data, err := ioutil.ReadFile(filepathToDownload)
	if err != nil {
		http.Error(w, "error reading file from disk", http.StatusBadRequest)
		return
	}

	// detect content type
	detectedContentType := http.DetectContentType(data)
	parts := strings.Split(detectedContentType, "/")

	if parts[0] != "image" {
		w.Header().Set("Content-Disposition", "attachment; filename="+filenameToDownload+"")
	} else {
		if download {
			// force download of image
			w.Header().Set("Content-Disposition", "attachment; filename="+filenameToDownload+"")
		}
	}

	logrus.Debugf("read %d bytes, filename: %s, content-type: %s", len(data), filenameToDownload, detectedContentType)
	w.Header().Set("Content-Type", detectedContentType)
	http.ServeContent(w, req, filenameToDownload, time.Now(), bytes.NewReader(data))
}

func UploadHandler(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:
		upload(w, req)
		return
	case http.MethodDelete:
		delete(w, req)
		return
	}
	errorMsg := fmt.Sprintf("Method [%s] is not supported:", req.Method)
	http.Error(w, errorMsg, http.StatusMethodNotAllowed)
}

func upload(w http.ResponseWriter, req *http.Request) {
	uuid := req.FormValue(paramUuid)
	if len(uuid) == 0 {
		logrus.Error("No uuid received, invalid upload request")
		http.Error(w, "No uuid received", http.StatusBadRequest)
		return
	}
	logrus.Infof("Starting upload handling of request with uuid of [%s]\n", uuid)
	file, headers, err := req.FormFile(paramFile)
	if err != nil {
		writeUploadResponse(w, err)
		return
	}

	fileDir := fmt.Sprintf("%s/%s", uploadDir, uuid)
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		writeUploadResponse(w, err)
		return
	}

	var filename string
	partIndex := req.FormValue(paramPartIndex)
	if len(partIndex) == 0 {
		filename = fmt.Sprintf("%s/%s", fileDir, headers.Filename)

	} else {
		filename = fmt.Sprintf("%s/%s_%05s", fileDir, uuid, partIndex)
	}
	outfile, err := os.Create(filename)
	if err != nil {
		writeUploadResponse(w, err)
		return
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, file)
	if err != nil {
		writeUploadResponse(w, err)
		return
	}

	writeUploadResponse(w, nil)
}

func delete(w http.ResponseWriter, req *http.Request) {
	logrus.Infof("Delete request received for uuid [%s]", req.URL.Path)
	err := os.RemoveAll(uploadDir + "/" + req.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

}

func ChunksDoneHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		errorMsg := fmt.Sprintf("Method [%s] is not supported", req.Method)
		http.Error(w, errorMsg, http.StatusMethodNotAllowed)
	}
	uuid := req.FormValue(paramUuid)
	filename := req.FormValue(paramFileName)
	totalFileSize, err := strconv.Atoi(req.FormValue(paramTotalFileSize))
	if err != nil {
		writeHttpResponse(w, http.StatusInternalServerError, err)
		return
	}
	totalParts, err := strconv.Atoi(req.FormValue(paramTotalParts))
	if err != nil {
		writeHttpResponse(w, http.StatusInternalServerError, err)
		return
	}

	finalFilename := fmt.Sprintf("%s/%s/%s", uploadDir, uuid, filename)
	f, err := os.Create(finalFilename)
	if err != nil {
		writeHttpResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer f.Close()

	var totalWritten int64
	for i := 0; i < totalParts; i++ {
		part := fmt.Sprintf("%[1]s/%[2]s/%[2]s_%05[3]d", uploadDir, uuid, i)
		partFile, err := os.Open(part)
		if err != nil {
			writeHttpResponse(w, http.StatusInternalServerError, err)
			return
		}
		written, err := io.Copy(f, partFile)
		if err != nil {
			writeHttpResponse(w, http.StatusInternalServerError, err)
			return
		}
		partFile.Close()
		totalWritten += written

		if err := os.Remove(part); err != nil {
			logrus.Errorf("Error: %v", err)
		}
	}

	if totalWritten != int64(totalFileSize) {
		errorMsg := fmt.Sprintf("Total file size mistmatch, expected %d bytes but actual is %d", totalFileSize, totalWritten)
		http.Error(w, errorMsg, http.StatusMethodNotAllowed)
	}
}

func writeHttpResponse(w http.ResponseWriter, httpCode int, err error) {
	w.WriteHeader(httpCode)
	if err != nil {
		logrus.Errorf("An error happened: %v", err)
		w.Write([]byte(err.Error()))
	}
}

func writeUploadResponse(w http.ResponseWriter, err error) {
	uploadResponse := new(UploadResponse)
	if err != nil {
		uploadResponse.Error = err.Error()
	} else {
		uploadResponse.Success = true
	}
	w.Header().Set("Content-Type", "text/plain")
	json.NewEncoder(w).Encode(uploadResponse)
}

func ListFilesHandler(w http.ResponseWriter, req *http.Request) {
	listOfFiles := make([]uploadedFile, 0)

	err := filepath.Walk(uploadDir, func(listedPath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			parts := strings.Split(listedPath, "/")
			if len(parts) != 3 {
				logrus.Errorf("[list] strange amount of parts detected, skipping file %q", listedPath)
				return nil
			}
			newFile := uploadedFile{
				UID:  parts[1],
				Name: info.Name(),
				Size: info.Size(),
			}
			logrus.Debugf("[list] adding file: %+v", newFile)
			listOfFiles = append(listOfFiles, newFile)
		}
		return nil
	})

	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.MarshalIndent(listOfFiles, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

type uploadedFile struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

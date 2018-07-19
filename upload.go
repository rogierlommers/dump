package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

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

	t := time.Now()
	datePrefix := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
	fileDir := fmt.Sprintf("%s/%s-%s", uploadDir, datePrefix, uuid)

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

func writeHttpResponse(w http.ResponseWriter, httpCode int, err error) {
	w.WriteHeader(httpCode)
	if err != nil {
		logrus.Errorf("An error happened: %v", err)
		w.Write([]byte(err.Error()))
	}
}

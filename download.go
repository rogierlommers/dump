package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type uploadedFile struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Size int64  `json:"size"`
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
	AddDownload(filenameToDownload, req.Referer(), fmt.Sprintf("%s", filepathToDownload))

	w.Header().Set("Content-Type", detectedContentType)
	http.ServeContent(w, req, filenameToDownload, time.Now(), bytes.NewReader(data))
}

func ListFilesHandler(w http.ResponseWriter, req *http.Request) {
	listOfFiles := make([]uploadedFile, 0)

	err := filepath.Walk(uploadDir, func(listedPath string, info os.FileInfo, err error) error {
		if !info.IsDir() {

			parts := strings.Split(filepath.Dir(listedPath), "/")
			extractedUID := parts[len(parts)-1]

			if len(extractedUID) < 10 {
				logrus.Errorf("strange uid detected, skipping file %q", listedPath)
				return nil
			}

			newFile := uploadedFile{
				UID:  extractedUID,
				Name: info.Name(),
				Size: info.Size(),
			}

			logrus.Debugf("adding file: %+v", newFile)
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

package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

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

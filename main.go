package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rogierlommers/dump/static"
	"github.com/sirupsen/logrus"
)

const (
	host = "0.0.0.0"
	port = 8080
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", static.IndexHandler)

	logrus.Infof("deamon running on host %s and port %d", host, port)

	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}

}

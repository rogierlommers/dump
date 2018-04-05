package static

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("running...")); err != nil {
		logrus.Error(err)
	}
}

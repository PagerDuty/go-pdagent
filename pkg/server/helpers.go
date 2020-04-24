package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func okResp(rw http.ResponseWriter, resp interface{}) {
	body, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(500)
		_, _ = fmt.Fprint(rw, err.Error())
		return
	}

	rw.WriteHeader(200)
	_, _ = rw.Write(body)
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

func errorResp(rw http.ResponseWriter, code int, errs []string) {
	body, err := json.Marshal(ErrorResponse{errs})
	if err != nil {
		rw.WriteHeader(500)
		_, _ = fmt.Fprint(rw, err.Error())
		return
	}

	rw.WriteHeader(code)
	_, _ = rw.Write(body)
}

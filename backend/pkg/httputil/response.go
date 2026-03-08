package httputil

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error *ErrorBody  `json:"error"`
	Meta  Meta        `json:"meta"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Timestamp int64 `json:"timestamp"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Data: data,
		Meta: Meta{Timestamp: time.Now().Unix()},
	})
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Error: &ErrorBody{Code: code, Message: message},
		Meta:  Meta{Timestamp: time.Now().Unix()},
	})
}

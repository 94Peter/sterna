package err

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiError interface {
	GetStatus() int
	GetErrorKey() string
	GetErrorMsg() string
	error
}

type myApiError struct {
	statusCode int
	message    string
	key        string
}

func (e myApiError) GetStatus() int {
	return e.statusCode
}

func (e myApiError) GetErrorKey() string {
	return e.key
}

func (e myApiError) GetErrorMsg() string {
	return e.message
}

func (e myApiError) Error() string {
	return fmt.Sprintf("%v: %v", e.statusCode, e.message)
}

func New(status int, msg string) ApiError {
	return myApiError{statusCode: status, message: msg}
}

func NewWithKey(status int, msg string, key string) ApiError {
	return myApiError{statusCode: status, message: msg, key: key}
}

func OutputErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if apiErr, ok := err.(ApiError); ok {
		outJson(w, apiErr.GetStatus(),
			map[string]interface{}{
				"status":   apiErr.GetStatus(),
				"title":    apiErr.GetErrorMsg(),
				"errorKey": apiErr.GetErrorKey(),
			})
	} else {
		outJson(w, http.StatusInternalServerError,
			map[string]interface{}{
				"status":   http.StatusInternalServerError,
				"title":    err.Error(),
				"errorKey": "",
			})
	}
	return
}

func outJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

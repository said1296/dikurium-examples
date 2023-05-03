package middleware

import "encoding/json"

type ContextKey struct {
	Name string
}

type httpError struct {
	Message string `json:"message"`
}

func errorJson(err error) string {
	httpErr := httpError{err.Error()}
	httpErrJson, _ := json.Marshal(httpErr)
	return string(httpErrJson)
}

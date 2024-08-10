package httpserverapi

import "net/http"

type Component interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

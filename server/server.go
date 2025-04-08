package server

import "github.com/gin-gonic/gin"

type Route struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

type Server interface {
	Start() error
	Port() int
	RegisterRoutes(routes []*Route)
}

package server

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	port   int
	router *gin.Engine
}

func NewHttpServer(port int) *HttpServer {
	// Create new gin server
	r := gin.Default()

	return &HttpServer{
		port:   port,
		router: r,
	}
}

func (s *HttpServer) Start() error {
  return s.router.Run(fmt.Sprintf(":%d", s.port))
}

func (s *HttpServer) RegisterRoutes(routes []*Route) {
  for _, route := range routes {
    s.router.Handle(route.Method, route.Path, route.Handler)
  }
}

func (s *HttpServer) Port() int {
  return s.port
}


func (s *HttpServer) EnableCors() {
  config := cors.DefaultConfig()
  config.AllowAllOrigins = true
  s.router.Use(cors.New(config))
}

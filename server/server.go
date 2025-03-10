package server

type Route struct {
	Method  string
	Path    string
	Handler interface{}
}

type Server interface {
	Start() error
	Port() int
	RegisterRoutes(routes []*Route)
}

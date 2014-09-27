package server

import "github.com/go-martini/martini"

// define as Server Options struct
type ServerOptions struct {
	MasterKey []byte
	StaticDir string
}

// define a Server interface abstraction
type Server interface {
	Start()
}

// create a private member containing the implementation
type privfilesMartini struct {
	*martini.ClassicMartini
}

// make ClassicMartini implment the server interface
func (m *privfilesMartini) Start() {
	m.Run()
}

// creates a new Server interface
func New(options ServerOptions) Server {
	if &options == nil || &options.MasterKey == nil || len(options.MasterKey) <= 0 {
		panic("FAILURE! ServerOptions.MasterKey is required to create a new server.")
	}

	m := martini.Classic()

	if &options.StaticDir != nil && len(options.StaticDir) > 0 {
		staticOptions := martini.StaticOptions{Prefix: "", SkipLogging: false, IndexFile: "index.html"}
		m.Use(martini.Static(options.StaticDir, staticOptions))
	}

	fileCtrl := FileController{options.MasterKey}
	m.Post("/upload", fileCtrl.Upload)
	m.Get("/download/:fileId/:mac", fileCtrl.Download)

	// return a privfilesMartini instance
	return &privfilesMartini{m}
}

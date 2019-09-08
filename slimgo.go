package slimgo

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

const Version = "slimgo v1.0.0"
const Release = "release"
const Debug = "debug"

// Server http server handler
type Server struct {
	router     *Router
	mode       string
	logger     Logger
	middleware []Handler
	lock       sync.Locker
}

// New create new server handler
// so we can use http.ListenAndServe()
func New() *Server {
	s := &Server{
		router: &Router{
			RedirectTrailingSlash: true,
			RedirectFixedPath:     true,
			NotFound:              defaultNotFoundHandler,
			MethodNotAllowed:      defaultMethodNotAllowHandler,
		},
		logger: &logger{
			debug: true,
		},
		mode: Debug,
	}

	fmt.Printf(banner, Version, runtime.Version())
	return s
}

func (s *Server) SetMode(mode string) {
	s.mode = mode
	s.logger = &logger{
		debug: mode == Debug,
	}
}

func (s *Server) SetLogger(l Logger) {
	s.logger = l
}

func (s *Server) Start(addr string) error {
	if err := http.ListenAndServe(addr, s); err != nil {
		return err
	}
	s.logger.Infof("web server listen on port %s (http)", addr)
	return nil
}

func (s *Server) StartTLS(addr, certFile, keyFile string) error {
	if err := http.ListenAndServeTLS(addr, certFile, keyFile, s); err != nil {
		return err
	}
	s.logger.Infof("web server listen on port %s (https)", addr)

	return nil
}

func (s *Server) Use(middleware ...Handler) {
	for _, filter := range middleware {
		s.middleware = append(s.middleware, filter)
		if s.mode != Release {
			fileName, file, line := GetFuncInfo(filter)
			fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
			_, file = path.Split(file)
			s.logger.Debugf("use middleware: {%s} ", fmt.Sprintf("%s[%d]:%s", file, line, fileName))
		}
	}
}

// Register registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (s *Server) Register(method string, relativePath string, handlers ...Handler) {
	if relativePath[0] != '/' {
		panic("path must begin with '/' in path '" + relativePath + "'")
	}

	if handlers == nil || len(handlers) == 0 {
		panic(fmt.Errorf("register [%s] %s failed: handler can not be null", method, relativePath))
	}

	relativePath = CleanURLPath(relativePath)
	// Update psMaxLen
	if pc := countParams(relativePath); pc > s.router.psMaxLen {
		s.router.psMaxLen = pc
	}

	if s.router.trees == nil {
		s.router.trees = make(map[string]*node)
	}

	root := s.router.trees[method]
	if root == nil {
		root = new(node)
		s.router.trees[method] = root
	}

	root.addRoute(relativePath, handlers)

	if s.mode == Debug {
		fileName, file, line := GetFuncInfo(handlers[len(handlers)-1])
		fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
		_, file = path.Split(file)
		s.logger.Debugf("mapped handler: %s %s {%s}",
			method, relativePath,
			fmt.Sprintf("%s[%d]:%s", file, line, fileName))
	}
}

// GET is a shortcut for router.Register("GET", path, handler)
func (s *Server) GET(relativePath string, handlers ...Handler) {
	s.Register("GET", relativePath, handlers...)
}

// HEAD is a shortcut for router.Register("HEAD", path, handler)
func (s *Server) HEAD(relativePath string, handlers ...Handler) {
	s.Register("HEAD", relativePath, handlers...)
}

// OPTIONS is a shortcut for router.Register("OPTIONS", path, handler)
func (s *Server) OPTIONS(relativePath string, handlers ...Handler) {
	s.Register("OPTIONS", relativePath, handlers...)
}

// POST is a shortcut for router.Register("POST", path, handler)
func (s *Server) POST(relativePath string, handlers ...Handler) {
	s.Register("POST", relativePath, handlers...)
}

// PUT is a shortcut for router.Register("PUT", path, handler)
func (s *Server) PUT(relativePath string, handlers ...Handler) {
	s.Register("PUT", relativePath, handlers...)
}

// PATCH is a shortcut for router.Register("PATCH", path, handler)
func (s *Server) PATCH(relativePath string, handlers ...Handler) {
	s.Register("PATCH", relativePath, handlers...)
}

// DELETE is a shortcut for router.Register("DELETE", path, handler)
func (s *Server) DELETE(relativePath string, handlers ...Handler) {
	s.Register("DELETE", relativePath, handlers...)
}

// 路由组
type groupRoutes struct {
	rootPath string
	filters  []Handler
	*Server
}

func (s *Server) Root(rootPath string, filters ...Handler) *groupRoutes {
	return &groupRoutes{
		rootPath: rootPath,
		filters:  filters,
		Server:   s,
	}
}

func (g *groupRoutes) AddRouterFilter(filters ...Handler) *groupRoutes {
	g.filters = append(g.filters, filters...)
	return g
}

func (g *groupRoutes) ClearRouterFilters() *groupRoutes {
	g.filters = g.filters[:0]
	return g
}

func (g *groupRoutes) combineHandlers(handlers ...Handler) []Handler {
	result := make([]Handler, len(g.filters)+len(handlers))
	copy(result, g.filters)
	copy(result[len(g.filters):], handlers)
	return result
}

func (g *groupRoutes) reg(method string, relativePath string, handlers ...Handler) *groupRoutes {
	relativePath = fmt.Sprintf("/%s/%s", g.rootPath, relativePath)
	g.Register(method, relativePath, g.combineHandlers(handlers...)...)
	return g
}

func (g *groupRoutes) GET(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("GET", relativePath, handlers...)
}

func (g *groupRoutes) HEAD(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("HEAD", relativePath, handlers...)
}

func (g *groupRoutes) OPTIONS(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("OPTIONS", relativePath, handlers...)
}

func (g *groupRoutes) POST(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("POST", relativePath, handlers...)
}

func (g *groupRoutes) PUT(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("PUT", relativePath, handlers...)
}

func (g *groupRoutes) PATCH(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("PATCH", relativePath, handlers...)
}

func (g *groupRoutes) DELETE(relativePath string, handlers ...Handler) *groupRoutes {
	return g.reg("DELETE", relativePath, handlers...)
}

// implement ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		// default panic handler
		if err := recover(); err != nil {
			// now this is raw response can not use compress
			w.Header().Del("Content-Encoding")
			defaultPanicHandler(w, req, err, debug.Stack())
			s.logger.Errorf("%s", debug.Stack())
		}
	}()

	w.Header().Add("Server", Version)
	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(400)
		return
	}

	c := newContext()
	c.init(s, w, req, s.middleware...)
	c.run()
	c.recycle()
}

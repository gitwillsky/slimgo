package server

import (
	"net/http"
	"fmt"
	"github.com/gitwillsky/slimgo/log"
	"runtime/debug"
	"strings"
	"github.com/gitwillsky/slimgo/utils"
	"time"
	"runtime"
	"path"
)

const Version = "slimgo1.0.1"

// Server http server handler
type Server struct {
	globalFilters []Handler
	router        *Router
	// NotFound Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound Handler
	// MethodNotAllowed Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed Handler
	servers          []*http.Server
	startTime        time.Time
}

// New create new server handler
// so we can use http.ListenAndServe()
func New() *Server {
	s := &Server{
		globalFilters: make([]Handler, 0),
		router: &Router{
			RedirectTrailingSlash: true,
			RedirectFixedPath:     true,
		},
		NotFound:         defaultNotFoundHandler,
		MethodNotAllowed: defaultMethodNotAllowHandler,
		servers:          make([]*http.Server, 0),
		startTime:        time.Now(),
	}

	fmt.Printf(banner, Version, runtime.Version())
	return s
}

func (s *Server) Start(addr string) error {
	log.Infof("web server listen on port %s (http)", addr)
	server := &http.Server{Addr: addr, Handler: s}
	s.servers = append(s.servers, server)
	log.Infof("started slimgo web application in %f seconds", time.Since(s.startTime).Seconds())
	return server.ListenAndServe()
}

func (s *Server) StartTLS(addr, certFile, keyFile string) error {
	log.Infof("web server listen on port %s (https)", addr)
	server := &http.Server{Addr: addr, Handler: s}
	s.servers = append(s.servers, server)
	log.Infof("started slimgo web application in %f seconds", time.Since(s.startTime).Seconds())
	return server.ListenAndServeTLS(certFile, keyFile)
}

// TODO graceFul shutdown
//func (s *Server) Shutdown() error {
//	for _, server := range s.servers {
//		server.Shutdown()
//	}
//}

// PreRouting 注册路由之前的过滤器
func (s *Server) AddFilter(filters ...Handler) {
	for _, filter := range filters {
		s.globalFilters = append(s.globalFilters, filter)
		if log.GetLevel() >= log.LevelInformational {
			for _, filter := range filters {
				fileName, file, line := utils.GetFuncInfo(filter)
				fileName = fileName[ strings.LastIndexByte(fileName, '.')+1:]
				_, file = path.Split(file)
				log.Infof("mapped filter: {%s} to: /*", fmt.Sprintf("%s[%d]:%s", file, line, fileName))
			}
		}
	}
}

// GET is a shortcut for router.Register("GET", path, handler)
func (s *Server) GET(urlPath string, handler Handler) {
	s.router.Register("GET", urlPath, handler)
}

// HEAD is a shortcut for router.Register("HEAD", path, handler)
func (s *Server) HEAD(urlPath string, handler Handler) {
	s.router.Register("HEAD", urlPath, handler)
}

// OPTIONS is a shortcut for router.Register("OPTIONS", path, handler)
func (s *Server) OPTIONS(urlPath string, handler Handler) {
	s.router.Register("OPTIONS", urlPath, handler)
}

// POST is a shortcut for router.Register("POST", path, handler)
func (s *Server) POST(urlPath string, handler Handler) {
	s.router.Register("POST", urlPath, handler)
}

// PUT is a shortcut for router.Register("PUT", path, handler)
func (s *Server) PUT(urlPath string, handler Handler) {
	s.router.Register("PUT", urlPath, handler)
}

// PATCH is a shortcut for router.Register("PATCH", path, handler)
func (s *Server) PATCH(urlPath string, handler Handler) {
	s.router.Register("PATCH", urlPath, handler)
}

// DELETE is a shortcut for router.Register("DELETE", path, handler)
func (s *Server) DELETE(urlPath string, handler Handler) {
	s.router.Register("DELETE", urlPath, handler)
}

// 路由组
type groupRoutes struct {
	rootPath string
	filters  []Handler
	server   *Server
}

func (s *Server) Root(rootPath string, handlers ...Handler) *groupRoutes {
	g := &groupRoutes{
		rootPath: rootPath,
		filters:  make([]Handler, len(handlers), (len(handlers)+1)*2),
		server:   s,
	}
	copy(g.filters, handlers)
	return g
}

func (g *groupRoutes) AddFilter(handlers ...Handler) *groupRoutes {
	g.filters = append(g.filters, handlers...)
	return g
}

func (g *groupRoutes) ClearFilters() *groupRoutes {
	g.filters = make([]Handler, 0)
	return g
}

func (g *groupRoutes) combineHandlers(handler Handler) []Handler {
	result := make([]Handler, len(g.filters)+1)
	copy(result, g.filters)
	result[len(g.filters)] = handler
	return result
}

func (g *groupRoutes) GET(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("GET", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) HEAD(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("HEAD", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) OPTIONS(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("OPTIONS", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) POST(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("POST", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) PUT(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("PUT", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) PATCH(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("PATCH", urlPath, g.combineHandlers(handler)...)
	return g
}

func (g *groupRoutes) DELETE(urlPath string, handler Handler) *groupRoutes {
	urlPath = fmt.Sprintf("/%s/%s", g.rootPath, urlPath)
	g.server.router.Register("DELETE", urlPath, g.combineHandlers(handler)...)
	return g
}

func (s *Server) initContext(ctx *Context) {
	urlPath := ctx.Request.URL.Path
	root, ok := s.router.trees[ctx.Request.Method]
	if !ok {
		ctx.AddHandlers(s.MethodNotAllowed)
		return
	}

	// get router handle
	regPath, handlers, params, tsr := root.getValue(urlPath)

	if handlers != nil {
		ctx.params = params
		ctx.regPath = regPath
		ctx.AddHandlers(handlers...)
		return
	}

	// fix path
	if ctx.Request.Method != "CONNECT" && urlPath != "/" {
		httpCode := 301 // 永久跳转
		if ctx.Request.Method != "GET" {
			httpCode = 307 // 暂时跳转
		}

		// need add/remove trailing slash
		if tsr && s.router.RedirectTrailingSlash {
			if len(urlPath) > 1 && urlPath[len(urlPath)-1] == '/' {
				// remove trailing slash
				ctx.Request.URL.Path = urlPath[:len(urlPath)-1]
			} else {
				// add trailing slash
				ctx.Request.URL.Path = urlPath + "/"
			}

			ctx.AddHandlers(func(context *Context) (interface{}, error) {
				http.Redirect(ctx.response, ctx.Request, ctx.Request.URL.String(), httpCode)
				return httpCode, nil
			})
			return
		}

		// maybe need clean path
		if s.router.RedirectFixedPath {
			cleanedPath, found := root.findCaseInsensitivePath(
				utils.CleanURLPath(urlPath),
				s.router.RedirectTrailingSlash,
			)
			if found {
				ctx.Request.URL.Path = utils.BytesToString(&cleanedPath)
				ctx.AddHandlers(func(context *Context) (interface{}, error) {
					http.Redirect(ctx.response, ctx.Request, ctx.Request.URL.String(), httpCode)
					return httpCode, nil
				})
			}
		}
	}

	// not found
	ctx.AddHandlers(s.NotFound)
}

// implement ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		// default panic handler
		if err := recover(); err != nil {
			// now this is raw response can not use compress
			w.Header().Del("Content-Encoding")
			defaultPanicHandler(w, req, err, debug.Stack())
			log.Error(debug.Stack())
		}
	}()

	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(400)
		return
	}

	w.Header().Add("Server", Version)

	c := newContext(w, req, s.globalFilters)
	s.initContext(c)

	r, e := c.Next()

	c.responseResolve(r, e)
	c.release()
}

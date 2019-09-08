package slimgo

import (
	"net/http"
	"sync"
)

// Router server router
type Router struct {
	// pool to recycle Param slices
	psPool sync.Pool

	// max number of params any of the path contains
	psMaxLen int

	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// NotFound Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound Handler
	// MethodNotAllowed Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed Handler
}

func (r *Router) psGet() *params {
	if ps := r.psPool.Get(); ps != nil {
		psp := ps.(*params)
		if cap(*psp) >= r.psMaxLen {
			*psp = (*psp)[0:0] // reset slice
			return psp
		}
	}

	// Allocate new slice if none is available
	ps := make(params, 0, r.psMaxLen)
	return &ps
}

func (r *Router) psRecycle(ps *params) {
	if ps != nil {
		r.psPool.Put(ps)
	}
}

func (r *Router) GetHandlers(req *http.Request) (handlers []Handler, regPath string, ps *params) {
	method := req.Method
	urlPath := req.URL.Path

	root, ok := r.trees[method]
	if !ok {
		handlers = append(handlers, r.MethodNotAllowed)
		return
	}

	ps = r.psGet()
	var tsr bool
	regPath, handlers, tsr = root.getValue(urlPath, ps)

	if len(handlers) > 0 {
		return
	}

	// fix path
	if method != "CONNECT" && urlPath != "/" {
		httpCode := 301 // 永久跳转
		if method != "GET" {
			httpCode = 307 // 暂时跳转
		}

		// need add/remove trailing slash
		if tsr && r.RedirectTrailingSlash {
			if len(urlPath) > 1 && urlPath[len(urlPath)-1] == '/' {
				// remove trailing slash
				urlPath = urlPath[:len(urlPath)-1]
			} else {
				// add trailing slash
				urlPath = urlPath + "/"
			}

			handlers = append(handlers, func(context Context) {
				http.Redirect(context.ResponseWriter(), context.Request(), urlPath, httpCode)
			})
			return
		}

		// maybe need clean path
		if r.RedirectFixedPath {
			cleanedPath, found := root.findCaseInsensitivePath(
				CleanURLPath(urlPath),
				r.RedirectTrailingSlash,
			)
			if found {
				urlPath = BytesToString(&cleanedPath)
				handlers = append(handlers, func(context Context) {
					http.Redirect(context.ResponseWriter(), context.Request(), urlPath, httpCode)
				})
				return
			}
		}
	}

	// not found
	handlers = append(handlers, r.NotFound)
	return
}

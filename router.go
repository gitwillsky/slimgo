package slimgo

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

// Router server router
type Router struct {
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
}

// Register registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Register(method, urlPath string, handlers ...Handler) {
	if urlPath[0] != '/' {
		Alert("path must begin with '/' in path '" + urlPath + "'")
	}

	if handlers == nil || len(handlers) == 0 {
		Alert("handlers can not be null")
	}

	urlPath = CleanURLPath(urlPath)

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(urlPath, handlers)

	if GetLevel() >= LevelInformational {
		buf := bytes.Buffer{}
		if len(handlers) > 1 {
			for index := 0; index < len(handlers)-1; index++ {
				if index > 0 {
					buf.WriteByte(',')
				}

				fileName, file, line := GetFuncInfo(handlers[index])
				fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
				_, file = path.Split(file)
				buf.WriteString(fmt.Sprintf("%s[%d]:%s", file, line, fileName))
			}

			Infof("mapped filter: {%s} to: [%s]%s", buf.String(), method, urlPath)
		}

		fileName, file, line := GetFuncInfo(handlers[len(handlers)-1])
		fileName = fileName[strings.LastIndexByte(fileName, '.')+1:]
		_, file = path.Split(file)
		Infof("mapped handler: {%s} to: [%s]%s", fmt.Sprintf("%s[%d]:%s", file, line, fileName), method, urlPath)
	}
}

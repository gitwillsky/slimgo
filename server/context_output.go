package server

import (
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/gitwillsky/slimgo/utils"
)

func (c *Context) responseResolve(data interface{}, e error) {
	if data == nil && e == nil {
		return
	}

	if e != nil {
		code := 500
		contentType := "text/plain; charset=utf-8"
		if err, ok := e.(*handlerError); ok {
			contentType = "application/json;charset=utf-8"
			code = err.StatusCode
		}
		c.response.Header().Set("Content-Type", contentType)
		c.response.WriteHeader(code)
		fmt.Fprintf(c.response, e.Error())
		return
	}

	if data != nil {
		reflectValue := reflect.ValueOf(data)
	WALK:
		typeKind := reflectValue.Type().Kind().String()
		typeName := reflectValue.Type().String()
		switch typeKind {
		case "string":
			fmt.Fprint(c.response, data)
		case "ptr":
			reflectValue = reflectValue.Elem()
			goto WALK
		case "struct":
			// file
			if typeName == "os.File" {
				file := data.(*os.File)
				d, err := file.Stat()
				if err != nil {
					http.Error(c.response, err.Error(), 500)
					return
				}
				http.ServeContent(c.response, c.Request, file.Name(), d.ModTime(), file)
				return
			}
			// convert struct to json and output response
			writeJson(c.response, data)
		case "map", "array", "slice":
			writeJson(c.response, data)
		case "int":
			if !c.wrote {
				c.WriteHeader(data.(int))
				return
			}
			fmt.Fprint(c.response, data)
		default:
			http.Error(c.response, "No resource resolver for "+typeName+", please add resouce filter to handle this", 500)
		}
	}
}

func writeJson(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json;charset=utf-8")
	dat, err := json.Marshal(data)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
	fmt.Fprint(res, utils.BytesToString(&dat))
}

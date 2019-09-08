package slimgo

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"runtime"
	"unsafe"
)

// GetFuncInfo 通过反射获取方法信息
func GetFuncInfo(i interface{}) (funcName, file string, line int) {
	pc := reflect.ValueOf(i).Pointer()
	funcInfo := runtime.FuncForPC(pc)
	funcName = funcInfo.Name()
	file, line = funcInfo.FileLine(pc)
	return
}

func minInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func BytesToString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

func StringToBytes(s *string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+" %s", val, suffix)
}

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}

// open or create log file.
func OpenORCreateFile(fullname string) (*os.File, error) {
	if len(fullname) > 0 {
		// prejudge dir exists.
		if _, err := os.Stat(fullname); os.IsNotExist(err) {
			// oh, it not exists, create.
			if err = os.Mkdir(path.Dir(fullname), 0666); os.IsNotExist(err) {
				panic("create file path failed, permission denied")
			}
		}

		// open file with write, append, create mode.
		if file, err := os.OpenFile(fullname, os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0666); err != nil {
			return nil, err
		} else {
			return file, nil
		}
	}
	return nil, errors.New("file name can not be null")
}

// internal helper to lazily create a buffer if necessary
func bufApp(buf *[]byte, s string, w int, c byte) {
	if *buf == nil {
		if s[w] == c {
			return
		}

		*buf = make([]byte, len(s))
		copy(*buf, s[:w])
	}
	(*buf)[w] = c
}

func CleanURLPath(urlPath string) string {
	// Turn empty string into "/"
	if urlPath == "" {
		return "/"
	}

	n := len(urlPath)
	var buf []byte

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1
	w := 1

	if urlPath[0] != '/' {
		r = 0
		buf = make([]byte, n+1)
		buf[0] = '/'
	}

	trailing := n > 2 && urlPath[n-1] == '/'

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp). So in contrast to the path package this
	// loop has no expensive function calls (except 1x make)

	for r < n {
		switch {
		case urlPath[r] == '/':
			// empty path element, trailing slash is added after the end
			r++

		case urlPath[r] == '.' && r+1 == n:
			trailing = true
			r++

		case urlPath[r] == '.' && urlPath[r+1] == '/':
			// . element
			r++

		case urlPath[r] == '.' && urlPath[r+1] == '.' && (r+2 == n || urlPath[r+2] == '/'):
			// .. element: remove to last /
			r += 2

			if w > 1 {
				// can backtrack
				w--

				if buf == nil {
					for w > 1 && urlPath[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}

		default:
			// real path element.
			// add slash if needed
			if w > 1 {
				bufApp(&buf, urlPath, w, '/')
				w++
			}

			// copy element
			for r < n && urlPath[r] != '/' {
				bufApp(&buf, urlPath, w, urlPath[r])
				w++
				r++
			}
		}
	}

	// re-append trailing slash
	if trailing && w > 1 {
		bufApp(&buf, urlPath, w, '/')
		w++
	}

	if buf == nil {
		return urlPath[:w]
	}
	return string(buf[:w])
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

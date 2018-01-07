package utils

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

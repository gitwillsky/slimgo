package server

import (
	"net/http"
	"fmt"
	"errors"
)

const banner = `
 ___  __    ____  __  __  ___  _____
/ __)(  )  (_  _)(  \/  )/ __)(  _  )
\__ \ )(__  _)(_  )    (( (_-. )(_)(
(___/(____)(____)(_/\/\_)\___/(_____)

Server Version: %s   Go Version: %s

`

const tpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Server Error</title>
    <style>
        * {padding: 0; margin: 0;}
        body{background: rgba(0,0,0,.05);line-height: 1.55;}
        header{width: 100%%;height: 8rem;color: #fff;}
        header>*{padding: 0 1.5rem;}
        header>h1{line-height: 5rem; background:#777; }
        header>p{background: #444; line-height: 2rem; text-align: left;}
        section{margin: 1rem 2rem 2rem 2rem;}
        section>h3{margin-bottom:10px;}
    </style>
</head>
<body>
    <article>
        <header>
            <h1>Server Error in Application</h1>
            <p>Server: slimgo v%s</p>
        </header>
       <section>
           <h3>Error Summary</h3>
           <p>HTTP Error %d - %s</p>
           <p>Request Method: %s</p>
           <p>Request URL: %s</p>
       </section>
       <section>
            <h3>What can I do?</h3>
            <h4>If you're a site visitor</h4>
            <p> Nothing you can do at the moment. If you need immediate assistance,
                please send us an email instead. We apologize for any inconvenience.</p>
            <h4>If you're the site owner</h4>
            <p>This error can only be fixed by server admins, please contact your website provider.</p>
       </section>
        <section>
            <h3>Debug Information</h3>
            <pre>%v</pre>
			<p>Stack Trace: </p>
			<pre>%s</pre>
        </section>
    </article>
</body>
</html>
`

// defaultPanicHandler default panic handler
func defaultPanicHandler(w http.ResponseWriter, req *http.Request, i interface{}, stack []byte) {
	returnHTML := fmt.Sprintf(tpl,
		Version, 500, http.StatusText(500), req.Method, req.URL.Path, i, stack)
	w.WriteHeader(500)
	fmt.Fprint(w, returnHTML)
}

// defaultNotFoundHandler default notfound handler
func defaultNotFoundHandler(ctx *Context) (interface{}, error) {
	return nil, ctx.NewError(404, errors.New("Not Found"))
}

// defaultMethodNotAllowHandler default method not allow handler
func defaultMethodNotAllowHandler(ctx *Context) (interface{}, error) {
	return nil, ctx.NewError(405, errors.New("Method Not Allowed"))
}


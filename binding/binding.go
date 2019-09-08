package binding

import "net/http"

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMEYAML              = "application/x-yaml"
)

// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	Name() string
	Bind(req *http.Request, target interface{}) error
}

// BindingBody adds BindBody method to Binding. BindBody is similar with Bind,
// but it reads the body from supplied bytes instead of req.Body.
type BindingBody interface {
	Binding
	BindBody(body []byte, target interface{}) error
}

// BindingUri adds BindUri method to Binding. BindUri is similar with Bind,
// but it read the Params.
type BindingUri interface {
	Binding
	BindUri(params map[string][]string, target interface{}) error
}

// StructValidator is the minimal interface which needs to be implemented in
// order for it to be used as the validator engine for ensuring the correctness
// of the request.
type StructValidator interface {
	// ValidateStruct can receive any kind of type and it should never panic, even if the configuration is not right.
	// If the received type is not a struct, any validation should be skipped and nil must be returned.
	// If the received type is a struct or pointer to a struct, the validation should be performed.
	// If the struct is not valid or the validation itself fails, a descriptive error should be returned.
	// Otherwise nil must be returned.
	ValidateStruct(interface{}) error

	// Engine returns the underlying validator engine which powers the
	// StructValidator implementation.
	Engine() interface{}
}

var Validator StructValidator = &defaultValidator{}

// Default returns the appropriate Binding instance based on the HTTP method
// and the content type.
func Default(method, contentType string) Binding {
	if method == "GET" {
		return &formBinding{}
	}

	switch contentType {
	case MIMEJSON:
		return &jsonBinding{}
	case MIMEMultipartPOSTForm:
		return &formMultipartBinding{}
	default: // case MIMEPOSTForm:
		return &formBinding{}
	}
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}

package binding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EnableDecoderUseNumber is used to call the UseNumber method on the JSON
// Decoder instance. UseNumber causes the Decoder to unmarshal a number into an
// interface{} as a Number instead of as a float64.
var EnableDecoderUseNumber = false

// EnableDecoderDisallowUnknownFields is used to call the DisallowUnknownFields method
// on the JSON Decoder instance. DisallowUnknownFields causes the Decoder to
// return an error when the destination is a struct and the input contains object
// keys which do not match any non-ignored, exported fields in the destination.
var EnableDecoderDisallowUnknownFields = false

type jsonBinding struct{}

var _ BindingBody = &jsonBinding{}

func (jsonBinding) Bind(req *http.Request, target interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	defer req.Body.Close()

	return decodeJSON(req.Body, target)
}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) BindBody(body []byte, target interface{}) error {
	return decodeJSON(bytes.NewReader(body), target)
}

func decodeJSON(r io.Reader, target interface{}) error {
	dec := json.NewDecoder(r)
	if EnableDecoderDisallowUnknownFields {
		dec.DisallowUnknownFields()
	}
	if EnableDecoderUseNumber {
		dec.UseNumber()
	}

	if err := dec.Decode(target); err != nil {
		return err
	}

	return validate(target)
}

package binding

import "github.com/asaskevich/govalidator"

type defaultValidator struct {
}

func (d *defaultValidator) ValidateStruct(target interface{}) error {
	_, err := govalidator.ValidateStruct(target)
	return err
}

func (d *defaultValidator) Engine() interface{} {
	return d
}

var _ StructValidator = &defaultValidator{}

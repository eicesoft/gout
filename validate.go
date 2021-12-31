package gout

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type sliceValidateError []error

func (err sliceValidateError) Error() string {
	var errMsgs []string
	for i, e := range err {
		if e == nil {
			continue
		}
		errMsgs = append(errMsgs, fmt.Sprintf("[%d]: %s", i, e.Error()))
	}
	return strings.Join(errMsgs, "\n")
}

type Validate struct {
	tagName string
}

func (v *Validate) Struct(obj interface{}) error {
	t := reflect.TypeOf(obj)
	for i := 0; i < t.NumField(); i++ {
		fmt.Printf("%s\n", t.Field(i).Tag)
	}

	return nil
}

func New() *Validate {
	v := &Validate{}
	return v
}

type defaultValidator struct {
	once     sync.Once
	validate *Validate
}

type StructValidator interface {
	ValidateStruct(interface{}) error
}

var Validator StructValidator = &defaultValidator{}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = New()
	})
}

func (v *defaultValidator) validateStruct(obj interface{}) error {
	v.lazyinit()
	return v.validate.Struct(obj)
}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(sliceValidateError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}

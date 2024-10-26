// Package validation is basically a wrapper around https://github.com/go-playground/validator and https://github.com/go-playground/mold.
//
// Besides the functionality of the go-playground packages it supports a more flexible error translation by using struct tags for error messages.
//
// The error message is set into the `errors` tag of the struct field, the key is the validator as provided in the `validate`
// struct tag. Multiple validators/errors are supported.
package validation

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strings"
)

type FieldErrors map[string]string

type tagErrors map[string]string

type Translate func(key string, args ...any) string

type Validator struct {
	backend *validator.Validate
}

var regexIndex = regexp.MustCompile(`(\[(\d+)])$`)

// NewValidator initializes and returns a new Validator instance with default configurations and custom validations.
func NewValidator() Validator {
	be := validator.New(validator.WithRequiredStructEnabled())
	_ = be.RegisterValidation("dateString", validateDateString)

	// use json field names, if defined, for reading the error tag
	be.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return Validator{
		backend: be,
	}
}

// RegisterCustomTypeFunc registers a custom validation function for specific types with the validator backend.
func (v Validator) RegisterCustomTypeFunc(fn validator.CustomTypeFunc, types ...interface{}) {
	v.backend.RegisterCustomTypeFunc(fn, types...)
}

// Var validates a single variable using the specified validation tag and returns a list of errors or nil if valid.
func (v Validator) Var(val any, tag string) ([]string, error) {
	err := v.backend.Var(val, tag)

	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return nil, err
	}

	errs := make([]string, 0)

	if err == nil {
		return errs, nil
	}

	for _, err := range err.(validator.ValidationErrors) {
		errs = append(errs, err.Tag())
	}

	return errs, nil
}

// Struct validates a given struct instance according to the validator rules set up and returns any field validation errors.
func (v Validator) Struct(s any, tl Translate) (FieldErrors, error) {
	fe := make(FieldErrors)

	t := reflect.TypeOf(s)

	err := v.backend.Struct(s)
	if err == nil {
		return fe, nil
	}

	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return fe, err
	}

	for _, err := range err.(validator.ValidationErrors) {

		ns := strings.Split(err.Namespace(), ".") // json field path, e.g. foo.bar
		ns = ns[1:]

		p := strings.Split(err.StructNamespace(), ".") // struct field name path, e.g. Foo.Bar
		p = p[1:]
		fld := getStructField(t, p)

		if fld != nil {
			errorsTag := fld.Tag.Get("errors")
			tags := parseErrorsTag(errorsTag)
			msg := tags[err.Tag()]

			if len(msg) == 0 {
				msg = err.Tag()
			}

			if tl != nil {
				msg = tl(msg)
			}

			// filter our anonymous fields, they are indicating embedded structs
			var filtered []string
			for i := range p {
				f := getStructField(t, p[:i+1])
				if f != nil && f.Anonymous == false {
					n := regexIndex.ReplaceAllString(ns[i], ".$2") // replace array index [x] with dot notation
					filtered = append(filtered, n)
				}
			}

			fe[strings.Join(filtered, ".")] = msg
		}
	}

	return fe, nil
}

// getStructField retrieves the specified field by following a path of names within a struct type using reflection.
func getStructField(s reflect.Type, path []string) *reflect.StructField {
	var field reflect.StructField

	t := s
	for _, p := range path {
		if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			t = t.Elem()
		}

		p = regexIndex.ReplaceAllString(p, "") // strip array index
		f, ok := t.FieldByName(p)
		if ok {
			field = f
			t = f.Type
		}
	}

	return &field
}

// parseErrorsTag parses the "errors" tag of a struct field into a map of error keys and their corresponding messages.
func parseErrorsTag(tag string) tagErrors {
	errMap := make(tagErrors)

	if len(tag) == 0 {
		return errMap
	}

	errs := strings.Split(tag, ",")

	for _, e := range errs {
		parts := strings.Split(e, ":")
		if len(parts) != 2 {
			continue
		}
		errMap[parts[0]] = parts[1]
	}

	return errMap
}

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

var regexIndex = regexp.MustCompile(`(\[\d])$`)

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
					n := regexIndex.ReplaceAllString(ns[i], "") // strip array index
					filtered = append(filtered, n)
				}
			}

			fe[strings.Join(filtered, ".")] = msg
		}
	}

	return fe, nil
}

func getStructField(s reflect.Type, path []string) *reflect.StructField {
	var field reflect.StructField

	t := s
	for _, p := range path {
		if t.Kind() == reflect.Ptr {
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

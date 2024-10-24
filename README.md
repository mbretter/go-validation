[![](https://github.com/mbretter/go-validation/actions/workflows/test.yml/badge.svg)](https://github.com/mbretter/go-validation/actions/workflows/test.yml)
[![](https://goreportcard.com/badge/mbretter/go-validation)](https://goreportcard.com/report/mbretter/go-validation "Go Report Card")
[![codecov](https://codecov.io/gh/mbretter/go-validation/graph/badge.svg?token=YMBMKY7W9X)](https://codecov.io/gh/mbretter/go-validation)

## Validating and sanitizing structs and variables

This is basically a wrapper around [go-playground/validator](https://github.com/go-playground/validator) and [go-playground/mold](https://github.com/go-playground/mold).

Besides the functionality of the go-playground packages it supports a more flexible error translation by using struct tags for error messages.

The error message is set into the `errors` tag of the struct field, the key is the validator as provided in the `validate` 
struct tag. Multiple validators/errors are supported.

## Validating structs

```go
type Address struct {
    Street string `json:"street" validate:"required" errors:"required:global.street.required"`
}

type LoginData struct {
	AppId    string  `json:"app_id" example:"E790D106-0C05-4263-882D-E5D665CF53C1"`
	Username string  `json:"username" validate:"required,email" errors:"required:global.username.required,email:global.username.invalid"`
	Password string  `json:"password"`
	Address  Address `json:"address"`
}

loginData := LoginData{
    Username: "",
    Address: Address{
        Street: "",
    },
}

fieldErrors, err := NewValidator().Struct(loginData, nil)
```

fieldErrors is a map containing the name of the failed field as key and the error message from the errors struct tag.
```
username -> global.username.required
address.street -> global.street.required
```
Nested fields are represented using a dot notation.

## Translation

In many cases the translation is application specific, it does not make sense integrating a sofisticated translation 
system into this package. However, it is possible to make the error translations by passing your own translation function.

The second argument to Struct() is an optional translation function with the type of `type Translate func(key string, args ...any) string`, 
this function accepts the text as given in the errors tag and should return the translated version of the error text.

If no translation func was given, the text as specified in the errors tag is returned unmodified.

```go

func Tr(key string, args ...any) string {
    // in the real world, do something useful
    return "tr." + key
}

fieldErrors, err := NewValidator().Struct(loginData, Tr)
```

## Slices of structs

The package supports slices or array of structs.

```go
type Address struct {
    Street string `json:"street" validate:"required" errors:"required:global.street.required"`
}

type Customer struct {
    Firstname string    `json:"firstname" validate:"required" errors:"required:global.firstname.required"`
    Lastname  string    `json:"lastname" validate:"required" errors:"required:global.lastname.required"`
    Addresses []Address `json:"addresses" validate:"dive"` // the dive keyword is important
}

addressOk := Address{Street: "Daham 66"}
addressWrong := Address{Street: ""}
data := Customer{Adresses: []Address{addressOk, addressWrong}}

fieldErrors, err := NewValidator().Struct(data, nil)
```

fieldErrors contains `addresses.1.street -> global.street.required`, the struct which contains the invalid 
value is indicated by its index.

This does work with slices of scalar values too
```go
type SliceOneOf struct {
    Name []string `json:"name" validate:"dive,oneof=one two three" errors:"oneof:name must be one two or three"`
}

data := SliceOneOf{Name: []string{"four"}}
fieldErrors, err := NewValidator().Struct(data, nil)
```

fieldErrors contains: `name.0 -> name must be one two or three`

## Sanitize

If you want to do some kind of sanitization, like triming, you can use the sanitizer which is a simple wrapper around 
go-playground/mold.

```go
type LoginData struct {
	AppId    string  `json:"app_id" mod:"trim"`
	Username string  `json:"username" mod:"trim,lcase"`
	Password string  `json:"password"`
	Address  Address `json:"address"`
}

loginData := LoginData2{
    AppId:    "E790D106-0C05-4263-882D-E5D665CF53C1 ",
    Username: "foo@bar.COM ",
}

err := NewSanitizer().Struct(&loginData)
```

In this case it trims the AppId field and the Username field, the Username field is modified to lowercase.
The loginData struct is modified in place.

For available modifiers see: [go-playground/mold](https://github.com/go-playground/mold).

## Examples

### Password quality check

```go
type RequestData struct {
    Username   string          `json:"username" mod:"trim,lcase" validate:"required,email" errors:"required:validation.global.email_required,email:validation.global.email_invalid"`
    Password1  string          `json:"password1" validate:"required,min=8,containsLowercase,containsUppercase,containsDigit,containsSpecialChar" errors:"required:validation.global.password_required,min:validation.global.password.minlength,containsLowercase:validation.global.password.lowercase,containsUppercase:validation.global.password.uppercase,containsDigit:validation.global.password.digit,containsSpecialChar:validation.global.password.special_character"`
    Password2  string          `json:"password2" validate:"required,eqfield=Password1" errors:"required:user.validation.password_required"`
}
```

### Validate non-empty fields only

```go
type RequestData struct {
    Id types.UUID    `json:"id" validate:"omitempty,uuid4" errors:"uuid4:validation.global.id_invalid"`
}
```

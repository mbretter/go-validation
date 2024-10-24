[![](https://github.com/mbretter/go-validation/actions/workflows/test.yml/badge.svg)](https://github.com/mbretter/go-validation/actions/workflows/test.yml)
[![](https://goreportcard.com/badge/mbretter/go-validation)](https://goreportcard.com/report/mbretter/go-validation "Go Report Card")
[![codecov](https://codecov.io/gh/mbretter/go-validation/graph/badge.svg?token=YMBMKY7W9X)](https://codecov.io/gh/mbretter/go-validation)

# validating and sanitizing structs and variables

This is basically a wrapper around [go-playground/validator](https://github.com/go-playground/validator) and [go-playground/mold](https://github.com/go-playground/mold).

Besides the functionality of the go-playground packages it supports a more flexible error translation by using struct tags for error messages.


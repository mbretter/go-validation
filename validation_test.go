package validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Address struct {
	Street string `json:"street" validate:"required" errors:"required:global.street.required"`
}

type LoginData struct {
	AppId    string  `json:"app_id" example:"E790D106-0C05-4263-882D-E5D665CF53C1"`
	Username string  `json:"username" format:"email" example:"foo@bar.com" validate:"required,email" errors:"required:global.username.required,email:global.username.invalid"`
	Password string  `json:"password"`
	Address  Address `json:"address"`
}

type Embedded struct {
	LoginData
	Status string `json:"status" validate:"oneof=active inactive" errors:"oneof:global.status.invalid"`
}

type Pointer struct {
	AppId    string  `json:"app_id" example:"E790D106-0C05-4263-882D-E5D665CF53C1"`
	Username *string `json:"username" format:"email" example:"foo@bar.com" validate:"email" errors:"email:global.username.invalid"`
	Password string  `json:"password"`
}

type PointerOmit struct {
	AppId *string `json:"app_id" validate:"omitnil,gt=0" errors:"gt:global.app_id.required" example:"E790D106-0C05-4263-882D-E5D665CF53C1"`
}

type NoJson struct {
	InternalId string `validate:"required"`
	Name       string `json:"-" validate:"required"`
}

type NoErrors struct {
	Name string `json:"name" validate:"required"`
}

type MalformedErrors struct {
	Name string `json:"name" validate:"required" errors:"required"`
}

type SliceOneOf struct {
	Name []string `json:"name" validate:"dive,oneof=one two three"`
}

type CustomValidators struct {
	Birthday string `json:"birthday" validate:"dateString" example:"1970-01-01"`
}

func Tr(key string, args ...any) string {
	return "tr." + key
}

func TestStruct_Ok(t *testing.T) {
	loginData := LoginData{
		Username: "foo@bar.com",
		Address: Address{
			Street: "Daham 66",
		},
	}

	fieldErrors, err := NewValidator().Struct(loginData, nil)

	assert.Nil(t, err)
	assert.Zero(t, len(fieldErrors))
}

func TestStruct_Fail(t *testing.T) {
	loginData := LoginData{
		Username: "",
		Address: Address{
			Street: "Daham 66",
		},
	}

	fieldErrors, err := NewValidator().Struct(loginData, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"username": "global.username.required"}, fieldErrors)
}

func TestStruct_PtrOk(t *testing.T) {
	loginData := LoginData{
		Username: "foo@bar.com",
		Address: Address{
			Street: "Daham 66",
		},
	}

	fieldErrors, err := NewValidator().Struct(&loginData, nil)

	assert.Nil(t, err)
	assert.Zero(t, len(fieldErrors))
}

func TestStruct_PtrInside_Fail(t *testing.T) {
	usr := "foo"
	loginData := Pointer{
		Username: &usr,
	}

	fieldErrors, err := NewValidator().Struct(&loginData, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"username": "global.username.invalid"}, fieldErrors)
}

func TestStruct_PtrOmit(t *testing.T) {
	appId := ""
	loginData := PointerOmit{
		AppId: &appId,
	}

	fieldErrors, err := NewValidator().Struct(loginData, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"app_id": "global.app_id.required"}, fieldErrors)

	loginData.AppId = nil

	fieldErrors, err = NewValidator().Struct(loginData, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{}, fieldErrors)
}

func TestStruct_ComposedFail(t *testing.T) {
	loginData := Embedded{
		LoginData: LoginData{
			Username: "foo@bar.com",
		},
		Status: "x",
	}

	fieldErrors, err := NewValidator().Struct(loginData, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"address.street": "global.street.required", "status": "global.status.invalid"}, fieldErrors)
}

func TestStruct_FailTranslate(t *testing.T) {
	loginData := LoginData{
		Username: "foo",
	}

	fieldErrors, err := NewValidator().Struct(loginData, Tr)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"address.street": "tr.global.street.required", "username": "tr.global.username.invalid"}, fieldErrors)
}

func TestStruct_NoJson(t *testing.T) {
	data := NoJson{}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"InternalId": "required", "Name": "required"}, fieldErrors)
}

func TestStruct_NoErrors(t *testing.T) {
	data := NoErrors{}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"name": "required"}, fieldErrors)
}

func TestStruct_MalformedErrors(t *testing.T) {
	data := MalformedErrors{}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"name": "required"}, fieldErrors)
}

func TestStruct_SliceOneOfSuccess(t *testing.T) {
	data := SliceOneOf{Name: []string{"one"}}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{}, fieldErrors)
}

func TestStruct_SliceOneOfFail(t *testing.T) {
	data := SliceOneOf{Name: []string{"four"}}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"name": "oneof"}, fieldErrors)
}

func TestVar_Ok(t *testing.T) {
	v := "foo"

	errs, err := NewValidator().Var(v, "required")

	assert.Nil(t, err)
	assert.Len(t, errs, 0)
}

func TestVar_Fail(t *testing.T) {
	v := ""

	errs, err := NewValidator().Var(v, "required")

	assert.Nil(t, err)
	assert.Equal(t, errs, []string{"required"})
}

func TestStruct_CustomValidators(t *testing.T) {
	data := CustomValidators{
		Birthday: "",
	}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Len(t, fieldErrors, 0)

	data.Birthday = "1930-01-01"
	fieldErrors, err = NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Len(t, fieldErrors, 0)
}

func TestStruct_CustomValidatorsFail(t *testing.T) {
	data := CustomValidators{
		Birthday: "x",
	}

	fieldErrors, err := NewValidator().Struct(data, nil)

	assert.Nil(t, err)
	assert.Equal(t, FieldErrors{"birthday": "dateString"}, fieldErrors)
}

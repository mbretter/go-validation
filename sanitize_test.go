package validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type LoginData2 struct {
	AppId    string  `json:"app_id" mod:"trim"`
	Username string  `json:"username" mod:"trim,lcase"`
	Password string  `json:"password"`
	Address  Address `json:"address"`
}

func TestSanitize_Ok(t *testing.T) {
	loginData := LoginData2{
		AppId:    "E790D106-0C05-4263-882D-E5D665CF53C1 ",
		Username: "foo@bar.COM ",
	}

	err := NewSanitizer().Struct(&loginData)

	assert.Nil(t, err)
	assert.Equal(t, LoginData2{
		AppId:    "E790D106-0C05-4263-882D-E5D665CF53C1",
		Username: "foo@bar.com",
	}, loginData)
}

func TestSanitize_Err(t *testing.T) {
	err := NewSanitizer().Struct(nil)

	if assert.NotNil(t, err) {
		assert.Equal(t, "mold: Struct(nil)", err.Error())
	}
}

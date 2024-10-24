package validation

import (
	"context"
	"github.com/go-playground/mold/v4"
	"github.com/go-playground/mold/v4/modifiers"
)

type Sanitizer struct {
	ctx     context.Context
	backend *mold.Transformer
}

func NewSanitizer() Sanitizer {
	be := modifiers.New()
	return Sanitizer{
		ctx:     context.Background(),
		backend: be,
	}
}

func (s Sanitizer) Struct(val any) error {
	err := s.backend.Struct(s.ctx, val)
	if err != nil {
		return err
	}
	return nil
}

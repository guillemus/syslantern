package validate

import (
	"fmt"
	"io"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
)

var V = validator.New(validator.WithRequiredStructEnabled())

func Unmarshal(payload io.Reader, v any) error {
	decoder := sonic.ConfigDefault.NewDecoder(payload)

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("decode strict json: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode strict json: unexpected trailing data")
		}
		return fmt.Errorf("decode strict json: %w", err)
	}

	if err := V.Struct(v); err != nil {
		return fmt.Errorf("validate strict json: %w", err)
	}

	return nil
}

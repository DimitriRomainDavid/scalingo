package controller

import (
	"errors"
	"scalingo/internal/core/domain"

	"github.com/go-playground/validator/v10"
)

func validateListProjects(input *domain.ListRepoInput) error {
	validate := validator.New()

	err := validate.Struct(input)
	if err != nil {
		return err
	}

	if (input.MinSize > input.MaxSize && input.MaxSize != 0) ||
		(input.MinSize == input.MaxSize && input.MinSize != 0) ||
		input.MinSize < 0 ||
		input.MaxSize < 0 {
		return errors.New("validation failed: max can't be less than min, min and max must be positives and different")
	}

	return nil
}

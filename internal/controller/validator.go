package controller

import (
	"errors"
	"scalingo/internal/core/domain"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-playground/validator/v10"
)

var validFields = map[string]bool{
	"language":             true,
	"license":              true,
	"name_contains":        true,
	"description_contains": true,
	"min_size":             true,
	"max_size":             true,
}

func validateListProjects(domainInput []byte) (*domain.ListRepoInput, error) {
	var input map[string]any
	if err := jsoniter.Unmarshal(domainInput, &input); err != nil {
		return nil, errors.New("List projects - unable to deserialize: " + err.Error())
	}

	for key := range input {
		if !validFields[key] {
			return nil, errors.New("invalid field: " + key)
		}
	}

	dInput := &domain.ListRepoInput{}

	err := jsoniter.Unmarshal(domainInput, &dInput)
	if err != nil {
		return nil, errors.New("List projects - unable to deserialize: " + err.Error())
	}

	validate := validator.New()

	err = validate.Struct(dInput)
	if err != nil {
		return nil, err
	}

	if (dInput.MinSize > dInput.MaxSize && dInput.MaxSize != 0) ||
		(dInput.MinSize == dInput.MaxSize && dInput.MinSize != 0) ||
		dInput.MinSize < 0 ||
		dInput.MaxSize < 0 {
		return nil, errors.New("validation failed: max can't be less than min, min and max must be positives and different")
	}
	return dInput, nil
}

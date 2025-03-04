package vapi

import (
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/modern-go/reflect2"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"reflect"
	"slices"
	"strings"
)

type DependencyItem struct {
	Name    string `mapstructure:"-"`
	Version string `mapstructure:"version"`
}

func ParseDependencies(deps map[string]any) ([]DependencyItem, error) {
	var dependencies []DependencyItem

	keys := maps.Keys(deps)
	slices.SortStableFunc(keys, strings.Compare)

	for name, depValue := range deps {
		record := DependencyItem{
			Name: name,
		}

		switch reflect2.TypeOf(depValue).Kind() {
		case reflect.String:
			record.Version = depValue.(string)
		case reflect.Map:
			if err := mapstructure.Decode(depValue, &record); err != nil {
				return nil, errors.Wrapf(err, "failed to decode dependency item")
			}
		default:
			return nil, errors.Wrapf(tclerrors.ErrValidation, "failed to parse dependency")
		}
		dependencies = append(dependencies, record)
	}

	return dependencies, nil
}

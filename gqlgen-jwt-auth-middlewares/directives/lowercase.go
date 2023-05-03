package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"strings"
)

func Lowercase(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	value, err := next(ctx)
	if err != nil {
		return nil, err
	}

	valueParsed, ok := value.(string)
	if ok {
		return strings.ToLower(valueParsed), nil
	}

	valueParsedPtr, ok := value.(*string)
	valueLowercase := strings.ToLower(*valueParsedPtr)
	if ok {
		return &valueLowercase, nil
	}

	return nil, errors.New("lowercase directive received a non-string value")
}
package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"strconv"
)

func IntBetween(ctx context.Context, obj interface{}, next graphql.Resolver, biggerThan *int, lessThan *int, fieldName string) (res interface{}, err error) {
	value, err := next(ctx)
	if biggerThan != nil && value.(int) <= *biggerThan {
		return nil, errors.New(fieldName + " value must be bigger than " + strconv.Itoa(*biggerThan))
	}

	if lessThan != nil && value.(int) >= *lessThan {
		return nil, errors.New(fieldName + "value must be less than " + strconv.Itoa(*lessThan))
	}

	return next(ctx)
}

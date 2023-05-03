package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"server/api/graphql/graph"
	"server/internal/user"
	"strings"
)

func Protected(ctx context.Context, obj interface{}, next graphql.Resolver, rules []graph.ProtectedRule) (res interface{}, err error) {
	userDB := GetLoggedUser(ctx)

	if userDB == nil  {
		return nil, ErrAuthorizationNoPermission.CompleteError(
			ctx,
			errors.New(graphql.GetFieldContext(ctx).Field.Name + " is protected, requires authentication with rules: " + rulesString(rules)),
			)
	}

	authorized := false

	for _, rule := range rules {
		switch rule {
		case graph.ProtectedRuleAdmin:
			if userDB.HasRole(user.RoleAdminID) {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		return nil, ErrAuthorizationNoPermission.CompleteError(
			ctx,
			errors.New(graphql.GetFieldContext(ctx).Field.Name + " field is protected by rules: " + rulesString(rules)),
		)
	}

	return next(ctx)
}

func rulesString(rules []graph.ProtectedRule) string {
	var rulesStringArr []string

	for _, rule := range rules {
		rulesStringArr = append(rulesStringArr, rule.String())
	}

	return strings.Join(rulesStringArr, ", ")
}
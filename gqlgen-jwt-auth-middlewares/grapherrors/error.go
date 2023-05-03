package grapherrors

import (
	"bytes"
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"strconv"
)

type Error struct {
	err        error
	Message    string                 `json:"message"`
	Path       ast.Path               `json:"path,omitempty"`
	Locations  []Location             `json:"locations,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Rule       string                 `json:"-"`
}

func NewError(code string) *Error {
	return &Error{
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

func (err *Error) CompleteError(ctx context.Context, e error) *gqlerror.Error {
	return &gqlerror.Error{
		Path:       graphql.GetPath(ctx),
		Message:    e.Error(),
		Extensions: map[string]interface{}{
			"code": err.Extensions["code"],
		},
	}
}

func (err *Error) SetFile(file string) {
	if file == "" {
		return
	}
	if err.Extensions == nil {
		err.Extensions = map[string]interface{}{}
	}

	err.Extensions["file"] = file
}

type Location struct {
	Line   int `json:"line,omitempty"`
	Column int `json:"column,omitempty"`
}

func (err *Error) Error() string {
	var res bytes.Buffer
	if err == nil {
		return ""
	}
	filename, _ := err.Extensions["file"].(string)
	if filename == "" {
		filename = "input"
	}
	res.WriteString(filename)

	if len(err.Locations) > 0 {
		res.WriteByte(':')
		res.WriteString(strconv.Itoa(err.Locations[0].Line))
	}

	res.WriteString(": ")
	if ps := err.pathString(); ps != "" {
		res.WriteString(ps)
		res.WriteByte(' ')
	}

	res.WriteString(err.Message)

	return res.String()
}

func (err Error) pathString() string {
	return err.Path.String()
}

func (err Error) Unwrap() error {
	return err.err
}

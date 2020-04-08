package lib

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	toValidation := func(mess ...string) Validation {
		return func() error {
			m := strings.Join(mess, " ")
			if m == "" {
				return nil
			}
			return fmt.Errorf(m)
		}
	}
	type args struct {
		validations []Validation
	}
	tests := []struct {
		name        string
		args        args
		expectedErr string
	}{
		{name: "test1", args: args{validations: []Validation{}}},
		{name: "test2", args: args{validations: []Validation{toValidation()}}},
		{name: "test3", args: args{validations: []Validation{toValidation("error1")}}, expectedErr: "error1"},
		{name: "test4", args: args{validations: []Validation{toValidation(), toValidation("error1")}}, expectedErr: "error1"},
		{name: "test5", args: args{validations: []Validation{toValidation("error0"), toValidation("error1")}}, expectedErr: "error0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.args.validations...)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

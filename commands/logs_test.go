package commands

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestValidateArgument(t *testing.T) {
	tests := []struct {
		name      string
		argName   string
		wantedVal string
		allVals   func() ([]string, error)
		wantErr   bool
	}{
		{
			name:      "valid argument",
			argName:   "something",
			wantedVal: "a",
			allVals: func() ([]string, error) {
				return []string{"a"}, nil
			},
			wantErr: false,
		},
		{
			name:      "not a valid argument",
			argName:   "something",
			wantedVal: "b",
			allVals: func() ([]string, error) {
				return []string{"a"}, nil
			},
			wantErr: true,
		},
		{
			name:      "allVals failure",
			argName:   "something",
			wantedVal: "b",
			allVals: func() ([]string, error) {
				return nil, fmt.Errorf("test")
			},
			wantErr: true,
		},
		{
			name:      "allVals nil content",
			argName:   "something",
			wantedVal: "b",
			allVals: func() ([]string, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:      "allVals empty content",
			argName:   "something",
			wantedVal: "b",
			allVals: func() ([]string, error) {
				return []string{}, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgument(tt.argName, tt.wantedVal, tt.allVals)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogCmdArguments(t *testing.T) {
	tests := []struct {
		name             string
		ctx              *components.Context
		wantErrMsgPrefix string
	}{
		{
			name: "zero argument  (without interactive menu)",
			ctx: &components.Context{
				Arguments: []string{},
			},
			wantErrMsgPrefix: "wrong number of arguments",
		},
		{
			name: "one argument",
			ctx: &components.Context{
				Arguments: []string{"a"},
			},
			wantErrMsgPrefix: "wrong number of arguments",
		},
		{
			name: "two argument",
			ctx: &components.Context{
				Arguments: []string{"a", "b"},
			},
			wantErrMsgPrefix: "wrong number of arguments",
		},
		{
			name: "three argument",
			ctx: &components.Context{
				Arguments: []string{"a", "b", "c"},
			},
			wantErrMsgPrefix: "server id",
		},
		{
			name: "four argument",
			ctx: &components.Context{
				Arguments: []string{"a", "b", "c", "d"},
			},
			wantErrMsgPrefix: "wrong number of arguments",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logsCmd(tt.ctx)
			assert.NotNil(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), tt.wantErrMsgPrefix))
		})
	}
}

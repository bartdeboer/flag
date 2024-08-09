package flag_test

import (
	"reflect"
	"testing"

	. "github.com/bartdeboer/flag"
)

func TestParseArguments(t *testing.T) {
	testCases := []struct {
		name             string
		args             []string
		expectedCommands []string
		expectedArgsMap  map[string]string
	}{
		{
			name:             "Single long arg with value",
			args:             []string{"--key=value"},
			expectedCommands: []string{},
			expectedArgsMap:  map[string]string{"key": "value"},
		},
		{
			name:             "Multiple commands",
			args:             []string{"command1", "command2"},
			expectedCommands: []string{"command1", "command2"},
			expectedArgsMap:  map[string]string{},
		},
		{
			name:             "Mixed args and commands",
			args:             []string{"cmd", "-k", "value", "--long", "other", "end"},
			expectedCommands: []string{"cmd", "end"},
			expectedArgsMap:  map[string]string{"k": "value", "long": "other"},
		},
		{
			name:             "Combined shorthand flags",
			args:             []string{"-abc"},
			expectedCommands: []string{},
			expectedArgsMap:  map[string]string{"a": "", "b": "", "c": ""},
		},
		{
			name:             "Long arg without value",
			args:             []string{"--key"},
			expectedCommands: []string{},
			expectedArgsMap:  map[string]string{"key": ""},
		},
		{
			name:             "Shorthand with value",
			args:             []string{"-k", "value"},
			expectedCommands: []string{},
			expectedArgsMap:  map[string]string{"k": "value"},
		},
		{
			name:             "Shorthand and long mix",
			args:             []string{"-k", "value", "--long=value2", "cmd", "--bool"},
			expectedCommands: []string{"cmd"},
			expectedArgsMap:  map[string]string{"k": "value", "long": "value2", "bool": ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commands, argsMap := ParseArguments(tc.args)
			if !reflect.DeepEqual(commands, tc.expectedCommands) {
				t.Errorf("Failed %s, Commands got: %v, want: %v", tc.name, commands, tc.expectedCommands)
			}
			if !reflect.DeepEqual(argsMap, tc.expectedArgsMap) {
				t.Errorf("Failed %s, ArgsMap got: %v, want: %v", tc.name, argsMap, tc.expectedArgsMap)
			}
		})
	}
}

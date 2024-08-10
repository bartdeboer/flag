package flag_test

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	. "github.com/bartdeboer/flag"
)

func TestPrintDefaults(t *testing.T) {
	type Config struct {
		PortNumber int    `usage:"Port to listen on" short:"p" default:"8080"`
		HostName   string `usage:"Host address" default:"localhost"`
		Verbose    bool   `usage:"Verbose mode" short:"v"`
		Timeout    *int   `usage:"Timeout in seconds" short:"t"`
	}
	testConfig := Config{}

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintDefaults(&testConfig)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	output := strings.TrimSpace(string(out))

	expected := `  -p --port-number int   Port to listen on (default 8080)
     --host-name string  Host address (default localhost)
  -v --verbose bool      Verbose mode
  -t --timeout *int      Timeout in seconds`

	if output != expected {
		t.Errorf("Expected output does not match actual output.\nExpected:\n%s\nActual:\n%s", expected, output)
	}
}

func TestParseSuccess(t *testing.T) {
	type Config struct {
		PortNumber int    `flag:"port" default:"8080"`
		AltPort    int    `default:"7070"`
		HostName   string `default:"localhost"`
		Restart    bool   `short:"r"`
		Verbose    bool   `flag:"verbose"`
		Slice      []string
	}
	args := []string{"--port=9090", "--host-name", "myserver.com", "--verbose=true", "-r", "--slice", "one,two,three"}

	var config Config

	if err := SetDefaults(&config); err != nil {
		t.Fatalf("SetDefaults failed: %v", err)
	}

	_, flags := ParseArgs(args)

	if err := SetFlags(&config, flags); err != nil {
		t.Fatalf("ParseArgs failed with error: %v", err)
	}

	if config.PortNumber != 9090 {
		t.Errorf("Expected port 9090, got %d", config.PortNumber)
	}
	if config.AltPort != 7070 {
		t.Errorf("Expected port 7070, got %d", config.AltPort)
	}
	if config.HostName != "myserver.com" {
		t.Errorf("Expected host 'myserver.com', got '%s'", config.HostName)
	}
	if !config.Verbose {
		t.Errorf("Expected verbose true, got %v", config.Verbose)
	}
	if !config.Restart {
		t.Errorf("Expected restart true, got %v", config.Restart)
	}
	if !reflect.DeepEqual(config.Slice, []string{"one", "two", "three"}) {
		t.Errorf("Expected slice 'one,two,three', got '%v'", config.Slice)
	}
}

func TestParseTypeError(t *testing.T) {
	type Config struct {
		Timeout int `flag:"timeout"`
	}
	args := []string{"--timeout=thirty"}

	var config Config

	// Redirect stdout to capture help output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_, flags := ParseArgs(args)
	err := SetFlags(&config, flags)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	if err == nil {
		t.Fatal("Expected error, got none")
	}

	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected help message to be printed")
	}

	expectedErrorMessage := "error parsing flag --timeout"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMessage, err.Error())
	}
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldType reflect.Type
		expected  interface{}
		expectErr bool
	}{
		{"int", "123", reflect.TypeOf(int(0)), int(123), false},
		{"int overflow", "99999999999999999999", reflect.TypeOf(int(0)), nil, true},
		{"bool true", "true", reflect.TypeOf(bool(false)), true, false},
		{"bool false", "false", reflect.TypeOf(bool(false)), false, false},
		{"bool invalid", "maybe", reflect.TypeOf(bool(false)), nil, true},
		{"string", "hello", reflect.TypeOf(""), "hello", false},
		{"float", "3.14159", reflect.TypeOf(float64(0)), 3.14159, false},
		{"float invalid", "pi", reflect.TypeOf(float64(0)), nil, true},
		{"slice strings", "one,two,three", reflect.TypeOf([]string{}), []string{"one", "two", "three"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var fieldVal reflect.Value
			if tc.fieldType.Kind() == reflect.Slice {
				fieldVal = reflect.New(tc.fieldType).Elem()
			} else {
				fieldVal = reflect.New(tc.fieldType).Elem()
			}

			err := SetField(fieldVal, tc.input, true)
			if (err != nil) != tc.expectErr {
				t.Errorf("setField() error = %v, expectErr %v", err, tc.expectErr)
			}

			if !tc.expectErr && !reflect.DeepEqual(fieldVal.Interface(), tc.expected) {
				t.Errorf("setField() got = %v, want %v", fieldVal.Interface(), tc.expected)
			}
		})
	}
}

func TestConfigParsing(t *testing.T) {
	type Config struct {
		PortNumber int    `env:"PORT" flag:"port" default:"8080"`
		HostName   string `default:"localhost"`
		LogLevel   string `default:"info"`
	}

	os.Setenv("PORT", "3000")
	os.Setenv("HOST_NAME", "example.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST_NAME")
	}()

	var config Config

	if err := SetDefaults(&config); err != nil {
		t.Fatalf("SetDefaults failed: %v", err)
	}

	if err := ParseEnv(&config); err != nil {
		t.Fatalf("ParseEnv failed: %v", err)
	}

	if config.PortNumber != 3000 {
		t.Errorf("Expected port to be 3000 from env, got %d", config.PortNumber)
	}
	if config.HostName != "example.com" {
		t.Errorf("Expected host to be 'example.com' from env, got '%s'", config.HostName)
	}
	if config.LogLevel != "info" {
		t.Errorf("Expected log level to be default 'info', got '%s'", config.LogLevel)
	}

	args := []string{"--log-level=debug"}

	_, flags := ParseArgs(args)
	err := SetFlags(&config, flags)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	if config.PortNumber != 3000 {
		t.Errorf("Port changed after parsing args, got %d", config.PortNumber)
	}
	if config.HostName != "example.com" {
		t.Errorf("Host changed after parsing args, got '%s'", config.HostName)
	}
	if config.LogLevel != "debug" {
		t.Errorf("Expected log level to be 'debug' from args, got '%s'", config.LogLevel)
	}
}

func TestParseAll(t *testing.T) {
	type Config struct {
		PortNumber int    `default:"8080"`
		HostName   string `default:"localhost"`
		LogLevel   string `default:"info"`
	}

	// Mock environment variables
	os.Setenv("PORT_NUMBER", "3000")
	os.Setenv("HOST_NAME", "example.com")
	defer func() {
		os.Unsetenv("PORT_NUMBER")
		os.Unsetenv("HOST_NAME")
	}()

	// Simulated command-line arguments
	args := []string{"--log-level=debug"}

	// Capture the output for --help flag
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var config Config
	remainingArgs, _, err := ParseAll(&config, append(args, "--help"))
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	// Check if help was printed
	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help message to be printed, got %s", output)
	}

	// Verify that the function exits after printing help
	if err != nil || remainingArgs != nil {
		t.Errorf("Expected no error and nil remainingArgs when help is printed, got error: %v, remainingArgs: %v", err, remainingArgs)
	}

	// Normal operation without --help
	remainingArgs, _, err = ParseAll(&config, args)
	if err != nil {
		t.Fatalf("SetAll failed: %v", err)
	}

	// Assertions to check if the values are as expected
	if config.PortNumber != 3000 { // Should come from environment variable
		t.Errorf("Expected Port to be 3000, got %d", config.PortNumber)
	}
	if config.HostName != "example.com" { // Should come from environment variable
		t.Errorf("Expected Host to be 'example.com', got '%s'", config.HostName)
	}
	if config.LogLevel != "debug" { // Should be set by command-line args
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", config.LogLevel)
	}

	// Check remaining arguments
	if len(remainingArgs) > 0 {
		t.Errorf("Expected no remaining arguments, got %v", remainingArgs)
	}
}

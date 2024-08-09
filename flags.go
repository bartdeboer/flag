package flag

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/bartdeboer/words"
)

// PrintDefaults generates a help page for the CLI based on struct tags with default values and types.
func PrintDefaults(config interface{}) {
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		fmt.Println("Expected a struct")
		return
	}

	typ := val.Type()
	maxNameTypeLength := 0
	entries := make([][3]string, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		usage := field.Tag.Get("usage")
		short := field.Tag.Get("short")
		def := field.Tag.Get("default")
		typeName := field.Type.Name()
		if field.Type.Kind() == reflect.Ptr {
			typeName = "*" + field.Type.Elem().Name()
		}

		// Constructing parts of the output
		shortPart := fmt.Sprintf("-%s", short)
		if short == "" {
			shortPart = "  " // Align when no shorthand is present
		}
		longPart := fmt.Sprintf("--%s %s", words.ToKebabCase(field.Name), typeName)
		defaultStr := ""
		if def != "" && def != "0" && def != "false" && def != "\"\"" {
			defaultStr = fmt.Sprintf(" (default %v)", def)
		}
		fullUsage := usage + defaultStr

		entry := longPart
		if len(entry) > maxNameTypeLength {
			maxNameTypeLength = len(entry)
		}
		entries[i] = [3]string{shortPart, entry, fullUsage}
	}

	fmt.Println("Usage:")
	for _, e := range entries {
		fmt.Printf("  %s %-*s  %s\n", e[0], maxNameTypeLength, e[1], e[2])
	}
}

// SetDefaults sets default values for fields in the config struct based on struct tags.
func SetDefaults(config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("config must be a pointer to a struct")
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue // Skip unexported fields
		}
		fieldType := t.Field(i)
		defaultValue := fieldType.Tag.Get("default")
		if defaultValue == "" {
			continue
		}

		err := SetField(field, defaultValue, false)
		if err != nil {
			return fmt.Errorf("error setting default for field %s: %v", fieldType.Name, err)
		}
	}
	return nil
}

// Parse parses the CLI arguments and populates the config struct.
func ParseArgs(config interface{}, args []string) ([]string, error) {
	outArgs, flags := ParseArguments(args)

	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("config must be a pointer to a struct")
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		var err error
		field := v.Field(i)
		fieldType := t.Field(i)
		shortName := fieldType.Tag.Get("short")
		flagName := fieldType.Tag.Get("flag")
		if flagName == "" {
			flagName = words.ToKebabCase(fieldType.Name)
		}
		if flagValue, exists := flags[shortName]; exists {
			err = SetField(field, flagValue, true)
		} else if flagValue, exists := flags[flagName]; exists {
			err = SetField(field, flagValue, true)
		}
		if err != nil {
			PrintDefaults(config) // Print help message
			return nil, fmt.Errorf("error parsing flag --%s: %v", flagName, err)
		}
	}

	return outArgs, nil
}

// SetField sets the field based on its type and the string value provided.
func SetField(field reflect.Value, value string, exists bool) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		if exists && value == "" {
			field.SetBool(true)
			return nil
		}
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Slice:
		// Assumes comma-separated values for slice types
		elemType := field.Type().Elem()
		if elemType.Kind() == reflect.String {
			field.Set(reflect.ValueOf(strings.Split(value, ",")))
		} else {
			// More complex parsing required for non-string slices
			return errors.New("complex slice types are not supported yet")
		}
	default:
		if field.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
			// Handle types that implement encoding.TextUnmarshaler
			unmarshaler := reflect.New(field.Type().Elem()).Interface().(encoding.TextUnmarshaler)
			if err := unmarshaler.UnmarshalText([]byte(value)); err != nil {
				return err
			}
			field.Set(reflect.ValueOf(unmarshaler).Elem())
		} else {
			return errors.New("unsupported flag type")
		}
	}
	return nil
}

// ParseEnv parses environment variables and populates the config struct.
func ParseEnv(config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("config must be a pointer to a struct")
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		envName := fieldType.Tag.Get("env")
		if envName == "" {
			envName = words.ToConstantCase(fieldType.Name)
		}

		envValue, exists := os.LookupEnv(envName)
		if !exists {
			continue // If environment variable is not set, skip setting the field
		}

		err := SetField(field, envValue, true)
		if err != nil {
			PrintDefaults(config) // Print help message if there's an error setting the field
			return fmt.Errorf("error setting environment variable %s: %v", envName, err)
		}
	}

	return nil
}

// ParseAll configures the application settings by setting defaults, parsing environment variables,
// and command-line arguments. It also checks for help flags (--help, -h) to display help messages.
func ParseAll(config interface{}, args []string) ([]string, error) {
	if err := SetDefaults(config); err != nil {
		return nil, fmt.Errorf("error setting default values: %v", err)
	}
	if err := ParseEnv(config); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %v", err)
	}
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			PrintDefaults(config)
			return nil, nil
		}
	}
	outArgs, err := ParseArgs(config, args)
	if err != nil {
		return nil, fmt.Errorf("error parsing command-line arguments: %v", err)
	}
	return outArgs, nil
}

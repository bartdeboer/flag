# Flag Package

The `flag` package provides utilities for creating CLI programs. It helps with parsing command-line arguments and environment variables, setting default values, and generating help messages based on struct tags.

## Installation

```sh
go get github.com/bartdeboer/flag
```

## Functions

### `PrintDefaults`

Prints the command-line help for all available flags defined within a struct. Each field of the struct should be tagged with usage (description of the flag) and optionally with short (short form of the flag).

```go
func PrintDefaults(config interface{})
```

Usage Example:

```go
type Config struct {
    Port int `short:"p" usage:"Port to listen on"`
    Host string `short:"h" usage:"Host address"`
}

var config Config
PrintDefaults(&config)
```

### `SetDefaults`

Sets default values for fields in a config struct based on default tags. This function is typically called before environment variables and command-line arguments are parsed.

```go
func SetDefaults(config interface{}) error
```

Usage Example:

```go
type Config struct {
    Port int `default:"8080"`
    Host string `default:"localhost"`
}

var config Config
err := SetDefaults(&config)
if err != nil {
    log.Fatalf("Error setting defaults: %v", err)
}
```

### `ParseEnv`

Parses environment variables and populates the config struct fields tagged with env. This function is usually called after setting default values and before parsing command-line arguments.

```go
func ParseEnv(config interface{}) error
```

Usage Example:

```go
type Config struct {
    PortNumber int // matches PORT_NUMBER
    HostName string `env:"APP_HOST"` // matches APP_HOST
}

var config Config
err := ParseEnv(&config)
if err != nil {
    log.Fatalf("Error parsing environment variables: %v", err)
}
```

### `SetFlags`

Parses command-line arguments and populates the config struct. Fields in the struct can be tagged with flag for long names and short for the abbreviated names. This function is usually called last to ensure it can override settings from defaults and environment variables.

```go
func SetFlags(config interface{}, flags map[string]string) error
```

Usage Example:

```go
type Config struct {
    PortNumber int `short:"p"` // matches --port-number and -p
    HostName string `flag:"host" short:"h"` // matches --host and -h
}

var config Config
args, flags := flag.ParseArgs(os.Args[1:])
err := SetFlags(&config, flags)
if err != nil {
    log.Fatalf("Error parsing command-line arguments: %v", err)
}
```

### `ParseAll`

Runs SetDefaults, ParseEnv and ParseArgs.

```go
func ParseAll(config interface{}, args []string) ([]string, map[string]string, error)
```

Usage Example:

```go
type Config struct {
    PortNumber int `short:"p"` // matches --port-number and -p
    HostName string `flag:"host" short:"h"` // matches --host and -h
}

var config Config
remainingArgs, flags, err := ParseAll(&config, os.Args[1:])
if err != nil {
    log.Fatalf("Error: %v", err)
}
```

## Getting Started

To use the flag package, define your configuration struct according to your application's requirements, annotate it with tags as described, and call these functions in the order of setting defaults, parsing environment variables, and finally parsing command-line arguments.

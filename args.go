package flag

import "strings"

// Parses out positional arguments, flags and shorthand flags from the slice
func ParseArgs(args []string) (positionalArgs []string, flags map[string]string) {
	positionalArgs = []string{}
	flags = make(map[string]string)

	i := 0
	for i < len(args) {
		arg := args[i]
		hasMoreArgs := i+1 < len(args)
		nextArgIsValue := hasMoreArgs && !strings.HasPrefix(args[i+1], "-")

		if strings.HasPrefix(arg, "--") {
			key := arg[2:]
			if strings.Contains(key, "=") {
				// Handle --key=value
				parts := strings.SplitN(key, "=", 2)
				flags[parts[0]] = parts[1]
			} else if nextArgIsValue {
				// Handle --key value
				flags[key] = args[i+1]
				i++ // Skip next arg as it's a value
			} else {
				// Handle --key
				flags[key] = ""
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			if len(arg) == 2 || strings.Contains(arg[2:], "=") {
				// Handle -k value or -k=value
				if strings.Contains(arg[2:], "=") {
					parts := strings.SplitN(arg[2:], "=", 2)
					flags[parts[0]] = parts[1]
				} else if nextArgIsValue {
					flags[arg[1:2]] = args[i+1]
					i++ // Skip next arg as it's a value
				} else {
					flags[arg[1:2]] = ""
				}
			} else {
				// Handle combined flags like -abc
				for _, flag := range arg[1:] {
					flags[string(flag)] = ""
				}
			}
		} else {
			// Positional arguments
			positionalArgs = append(positionalArgs, arg)
		}
		i++
	}

	return positionalArgs, flags
}

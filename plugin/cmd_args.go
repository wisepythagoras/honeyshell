package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

type OptConfig map[string]bool

type CmdArgs struct {
	RawArgs string
	argMap  map[string]any
}

func (args *CmdArgs) ParseOpts(optsConfig OptConfig) (map[string]any, []string, error) {
	parts := strings.Split(args.RawArgs, " ")
	bareArgs := make([]string, 0)
	opts := make(map[string]any)
	bypassNext := false

	for i, part := range parts {
		if part == "" || bypassNext {
			bypassNext = false
			continue
		}

		if hasArg, ok := optsConfig[part]; ok {
			// Check for arguments.
			if !hasArg {
				opts[strings.ReplaceAll(part, "-", "")] = true
			} else {
				if i >= len(parts)-1 || parts[i+1][0] == '-' {
					noValFound := fmt.Errorf("Argument %q requires a value, but none was found", part)
					return opts, bareArgs, noValFound
				}

				opts[strings.ReplaceAll(part, "-", "")] = parts[i+1]
				bypassNext = true
			}
		} else {
			bareArgs = append(bareArgs, part)
		}
	}

	return opts, bareArgs, nil
}

func (args *CmdArgs) Parse() {
	args.argMap = make(map[string]any)
	parts := strings.Split(args.RawArgs, " ")

	for i, part := range parts {
		part = strings.Trim(part, " ")

		if len(part) == 0 {
			continue
		}

		if part[0] == '-' && part[1] == '-' {
			key := strings.Trim(part, "-")

			if i < len(parts)-1 && parts[i+1][0] != '-' {
				args.argMap[key] = parts[i+1]
			}

			args.argMap[key] = true
		} else if part[0] == '-' {
			key := strings.Trim(part, "-")

			if i < len(parts)-1 && len(parts[i+1]) > 0 && parts[i+1][0] != '-' {
				args.argMap[key] = parts[i+1]
			}

			args.argMap[key] = true

			for _, char := range key {
				args.argMap[string(char)] = true
			}
		} else {
			args.argMap["raw"] = part
		}
	}
}

func (args *CmdArgs) Get(key string) any {
	if v, ok := args.argMap[key]; ok {
		return v
	}

	return nil
}

func (args *CmdArgs) Array() []string {
	re := regexp.MustCompile(`(\s+)`)
	rawArgs := strings.Trim(re.ReplaceAllString(args.RawArgs, " "), " ")
	return strings.Split(rawArgs, " ")
}

func (args *CmdArgs) ArrayWithCommand(cmd string) []string {
	cmdArgs := []string{cmd}
	return append(cmdArgs, args.Array()...)
}

func (args *CmdArgs) ForEach(callback func(string, any)) {
	for k, v := range args.argMap {
		callback(k, v)
	}
}

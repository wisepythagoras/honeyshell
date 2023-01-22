package plugin

import "strings"

type CmdArgs struct {
	RawArgs string
	argMap  map[string]any
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

			if i < len(parts)-1 && parts[i+1][0] != '-' {
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

func (args *CmdArgs) ForEach(callback func(string, any)) {
	for k, v := range args.argMap {
		callback(k, v)
	}
}

package plugin

import (
	"fmt"
	"regexp"
	"strings"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type OptConfig struct {
	Opts       map[string]bool
	ParsedOpts map[string]any
	BareArgs   []string
}

func (o *OptConfig) AddOne(arg string, hasArg bool) {
	o.Opts[arg] = hasArg
}

func (o *OptConfig) AddBoth(short, long string, hasArg bool) {
	o.Opts[short] = hasArg
	o.Opts[long] = hasArg
}

func (o *OptConfig) Get(arg string) any {
	if o.ParsedOpts == nil {
		return nil
	}

	if val, ok := o.ParsedOpts[arg]; ok {
		return val
	}

	return nil
}

type CmdArgs struct {
	RawArgs string
	argMap  map[string]any
}

func (args *CmdArgs) ParseOpts(optsConfig *OptConfig) (map[string]any, []string, error) {
	parts := strings.Split(args.RawArgs, " ")
	bareArgs := make([]string, 0)
	opts := make(map[string]any)
	bypassNext := false

	optsConfig.ParsedOpts = opts
	optsConfig.BareArgs = bareArgs

	for i, part := range parts {
		if part == "" || bypassNext {
			bypassNext = false
			continue
		}

		if hasArg, ok := optsConfig.Opts[part]; ok {
			// Check for arguments.
			if !hasArg {
				optsConfig.ParsedOpts[strings.ReplaceAll(part, "-", "")] = true
			} else {
				if i >= len(parts)-1 || parts[i+1][0] == '-' {
					noValFound := fmt.Errorf("Argument %q requires a value, but none was found", part)
					return optsConfig.ParsedOpts, optsConfig.BareArgs, noValFound
				}

				optsConfig.ParsedOpts[strings.ReplaceAll(part, "-", "")] = parts[i+1]
				bypassNext = true
			}
		} else {
			optsConfig.BareArgs = append(optsConfig.BareArgs, part)
		}
	}

	return optsConfig.ParsedOpts, optsConfig.BareArgs, nil
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

func CreateOptsConfig() *OptConfig {
	return &OptConfig{
		Opts: make(map[string]bool),
	}
}

func OptsModule(L *lua.LState) *lua.LTable {
	module := L.NewTable()

	L.SetField(module, "CreateOptsConfig", luar.New(L, CreateOptsConfig))

	return module
}

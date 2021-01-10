package config

import "flag"

type FlagStruct struct {
	FlagType     int
	Value        interface{}
	DefaultValue interface{}
	Usage        string
}

type Flag map[string]FlagStruct

const (
	FlagInt = iota
	FlagBool
	FlagString
)

func Parse(flags Flag) {
	for name, f := range flags {
		switch f.FlagType {
		case FlagInt:
			f.Value = flag.Int(name, f.DefaultValue.(int), f.Usage)
		case FlagBool:
			f.Value = flag.Bool(name, f.DefaultValue.(bool), f.Usage)
		case FlagString:
			f.Value = flag.String(name, f.DefaultValue.(string), f.Usage)
		}
		flags[name] = f
	}

	flag.Parse()
}

func (f Flag) Bool(c string) bool {
	if f, ok := f[c]; ok {
		return castToBool(f.Value)
	}
	return false
}

func (f Flag) String(c string) string {
	if f, ok := f[c]; ok {
		return castToString(f.Value)
	}

	return ""
}

func castToString(c interface{}) string {
	switch v := c.(type) {
	case *string:
		return *v
	}
	return ""
}

func castToBool(c interface{}) bool {
	switch v := c.(type) {
	case *bool:
		return *v
	}
	return false
}

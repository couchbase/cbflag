package cbflag

import (
	"fmt"
	"strings"
)

const OVERHEAD_LEN = 10
const FLAGS_LEN int = 20
const USAGE_LEN int = 50
const TOTAL_LEN int = 80

type ValidatorFn func(Value) error

type Flag struct {
	short      string
	long       string
	desc       string
	value      Value
	validator  ValidatorFn
	foundLong  bool
	foundShort bool
	required   bool
	isFlag     bool
}

func BoolFlag(result *bool, def bool, short, long, usage string) *Flag {
	return varFlag(newBoolValue(def, result), short, long, usage, nil, false, true)
}

func Float64Flag(result *float64, def float64, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newFloat64Value(def, result), short, long, usage, validator, required, false)
}

func IntFlag(result *int, def int, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newIntValue(def, result), short, long, usage, validator, required, false)
}

func Int64Flag(result *int64, def int64, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newInt64Value(def, result), short, long, usage, validator, required, false)
}

func StringFlag(result *string, def, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newStringValue(def, result), short, long, usage, validator, required, false)
}

func UintFlag(result *uint, def uint, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newUintValue(def, result), short, long, usage, validator, required, false)
}

func Uint64Flag(result *uint64, def uint64, short, long, usage string, validator ValidatorFn, required bool) *Flag {
	return varFlag(newUint64Value(def, result), short, long, usage, validator, required, false)
}

func HostFlag(result *string, def string, required bool) *Flag {
	return varFlag(newStringValue(def, result), "c", "cluster", "The hostname of the Couchbase cluster",
		hostValidator, required, false)
}

func UsernameFlag(result *string, def string, required bool) *Flag {
	return varFlag(newStringValue(def, result), "u", "username", "The hostname of the Couchbase cluster",
		nil, required, false)
}

func PasswordFlag(result *string, def string, required bool) *Flag {
	return varFlag(newStringValue(def, result), "p", "password", "The password of the Couchbase cluster",
		nil, required, false)
}

func CACertFlag(result *string, def string, required bool) *Flag {
	return varFlag(newStringValue(def, result), "p", "cacert",
		"Verifies the cluster identity with this certificate", nil, required, false)
}

func NoSSLVerifyFlag(result *bool, required bool) *Flag {
	return varFlag(newBoolValue(false, result), "n", "no-ssl-verify",
		"Skips SSL verification of certificates against CA", nil, required, true)
}

func helpFlag(result *bool) *Flag {
	return varFlag(newBoolValue(false, result), "h", "help", "Prints the help message", nil, false, true)
}

func varFlag(value Value, short, long, usage string, validator ValidatorFn, required, isFlag bool) *Flag {
	return &Flag{
		short:      short,
		long:       long,
		desc:       usage,
		value:      value,
		validator:  validator,
		foundLong:  false,
		foundShort: false,
		required:   required,
		isFlag:     isFlag,
	}
}

func (f *Flag) found() bool {
	return f.foundLong || f.foundShort
}

func (f *Flag) markFound(value string) {
	if strings.HasPrefix(value, "--") {
		f.foundLong = true
	} else if strings.HasPrefix(value, "-") {
		f.foundShort = true
	}
}

func (f *Flag) validate() error {
	if f.validator == nil {
		return nil
	}

	return f.validator(f.value)
}

func (f *Flag) usageString() string {
	s := ""
	lines := f.splitDescription()

	if len(f.long) > FLAGS_LEN {
		s += fmt.Sprintf("  -%s,--%s\n", f.short, f.long)
		s += fmt.Sprintf("%s%s\n", strings.Repeat(" ", TOTAL_LEN-USAGE_LEN), lines[0])
	} else {
		spaces := strings.Repeat(" ", TOTAL_LEN-USAGE_LEN-OVERHEAD_LEN-len(f.long))
		s += fmt.Sprintf("  -%s,--%s%s   %s\n", f.short, f.long, spaces, lines[0])
	}

	for i := 1; i < len(lines); i++ {
		s += fmt.Sprintf("%s%s\n", strings.Repeat(" ", TOTAL_LEN-USAGE_LEN), lines[i])
	}

	return s
}

func (f *Flag) splitDescription() []string {
	desc := f.desc
	lines := make([]string, 0)
	for len(desc) > 50 {
		i := 50
		for ; i >= 0; i-- {
			if desc[i] == ' ' {
				break
			}
		}
		lines = append(lines, desc[0:i])
		desc = desc[i+1:]
	}

	return append(lines, desc)
}

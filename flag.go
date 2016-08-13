package cbflag

import (
	"fmt"
	"strings"

	"github.com/couchbase/cbflag/pwd"
)

// |                                 TOTAL_LEN (80)                                 |
// | PREFIX_LEN (2) |       FLAGS_LEN (25)       | POSTFIX_LEN (3) | USAGE_LEN (50) |
// | PREFIX_LEN (2) | flags definition | padding | POSTFIX_LEN (3) | USAGE_LEN (50) |

const PREFIX_LEN = 2
const POSTFIX_LEN = 3
const FLAGS_LEN int = 25
const USAGE_LEN int = 50
const TOTAL_LEN int = 80

type ValidatorFn func(Value) error
type OptionHandler func(string, string) (string, bool, error)

type Flag struct {
	short      string
	long       string
	desc       string
	value      Value
	validator  ValidatorFn
	optHandler OptionHandler
	foundLong  bool
	foundShort bool
	required   bool
	hidden     bool
	isFlag     bool
}

func BoolFlag(result *bool, def bool, short, long, usage string, hidden bool) *Flag {
	return varFlag(newBoolValue(def, result), short, long, usage, nil, DefaultOptionHandler,
		false, hidden, true)
}

func Float64Flag(result *float64, def float64, short, long, usage string, validator ValidatorFn, required,
	hidden bool) *Flag {
	return varFlag(newFloat64Value(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func IntFlag(result *int, def int, short, long, usage string, validator ValidatorFn, required, hidden bool) *Flag {
	return varFlag(newIntValue(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func Int64Flag(result *int64, def int64, short, long, usage string, validator ValidatorFn, required,
	hidden bool) *Flag {
	return varFlag(newInt64Value(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func StringFlag(result *string, def, short, long, usage string, validator ValidatorFn, required,
	hidden bool) *Flag {
	return varFlag(newStringValue(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func UintFlag(result *uint, def uint, short, long, usage string, validator ValidatorFn, required,
	hidden bool) *Flag {
	return varFlag(newUintValue(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func Uint64Flag(result *uint64, def uint64, short, long, usage string, validator ValidatorFn,
	required, hidden bool) *Flag {
	return varFlag(newUint64Value(def, result), short, long, usage, validator, DefaultOptionHandler,
		required, hidden, false)
}

func HostFlag(result *string, def string, required, hidden bool) *Flag {
	return varFlag(newStringValue(def, result), "c", "cluster", "The hostname of the Couchbase cluster",
		HostValidator, DefaultOptionHandler, required, hidden, false)
}

func UsernameFlag(result *string, def string, required, hidden bool) *Flag {
	return varFlag(newStringValue(def, result), "u", "username", "The hostname of the Couchbase cluster",
		nil, DefaultOptionHandler, required, hidden, false)
}

func PasswordFlag(result *string, def string, required, hidden bool) *Flag {
	return varFlag(newStringValue(def, result), "p", "password", "The password of the Couchbase cluster",
		nil, PasswordOptionHandler, required, hidden, false)
}

func CACertFlag(result *string, def string, required, hidden bool) *Flag {
	return varFlag(newStringValue(def, result), "", "cacert",
		"Verifies the cluster identity with this certificate", nil, DefaultOptionHandler,
		required, hidden, false)
}

func NoSSLVerifyFlag(result *bool, required, hidden bool) *Flag {
	return varFlag(newBoolValue(false, result), "", "no-ssl-verify",
		"Skips SSL verification of certificates against CA", nil, DefaultOptionHandler,
		required, hidden, true)
}

func helpFlag(result *bool) *Flag {
	return varFlag(newBoolValue(false, result), "h", "help", "Prints the help message", nil,
		DefaultOptionHandler, false, false, true)
}

func varFlag(value Value, short, long, usage string, validator ValidatorFn, optHandler OptionHandler,
	required, hidden, isFlag bool) *Flag {
	return &Flag{
		short:      short,
		long:       long,
		desc:       usage,
		value:      value,
		validator:  validator,
		optHandler: optHandler,
		foundLong:  false,
		foundShort: false,
		required:   required,
		hidden:     hidden,
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
	if f.hidden {
		return ""
	}

	s := ""
	lines := f.splitDescription()

	flagsStr := f.flagsHelpString()
	if len(f.long) > FLAGS_LEN {
		prePadding := strings.Repeat(" ", PREFIX_LEN)
		s += fmt.Sprintf("%s%s\n", prePadding, flagsStr)
		s += fmt.Sprintf("%s%s\n", strings.Repeat(" ", TOTAL_LEN-USAGE_LEN), lines[0])
	} else {
		prePadding := strings.Repeat(" ", PREFIX_LEN)
		postPadding := strings.Repeat(" ", FLAGS_LEN+POSTFIX_LEN-len(flagsStr))
		s += fmt.Sprintf("%s%s%s%s\n", prePadding, flagsStr, postPadding, lines[0])
	}

	for i := 1; i < len(lines); i++ {
		s += fmt.Sprintf("%s%s\n", strings.Repeat(" ", TOTAL_LEN-USAGE_LEN), lines[i])
	}

	return s
}

func (f *Flag) flagsHelpString() string {
	if f.short != "" && f.long != "" {
		return fmt.Sprintf("-%s,--%s", f.short, f.long)
	} else if f.short == "" {
		return fmt.Sprintf("   --%s", f.long)
	} else if f.long == "" {
		return fmt.Sprintf("-%s", f.short)
	} else {
		return ""
	}
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

func DefaultOptionHandler(opt, value string) (string, bool, error) {
	if value == "" || strings.HasPrefix(value, "-") {
		return value, false, fmt.Errorf("Expected argument for option: %s", opt)
	}

	return value, true, nil
}

func PasswordOptionHandler(opt, value string) (string, bool, error) {
	if value == "" || strings.HasPrefix(value, "-") {
		fmt.Print("Password: ")
		password, err := pwd.GetPasswd()
		return string(password), false, err
	}

	return value, true, nil
}

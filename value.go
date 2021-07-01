/*
Copyright 2016-Present Couchbase, Inc.

Use of this software is governed by the Business Source License included in
the file licenses/BSL-Couchbase.txt.  As of the Change Date specified in that
file, in accordance with the Business Source License, use of this software will
be governed by the Apache License, Version 2.0, included in the file
licenses/APL2.txt.
*/

package cbflag

import ()

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *boolValue) IsBoolFlag() bool { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i) }

// -- rune Value
type runeValue rune

func newRuneValue(val rune, p *rune) *runeValue {
	*p = val
	return (*runeValue)(p)
}

func (s *runeValue) Set(val string) error {
	if len(val) == 0 {
		return errors.New("No value specified")
	} else if len(val) == 1 {
		*s = runeValue(val[0])
	} else if len(val) == 2 {
		if val[0] == '\\' && val[1] == 'n' {
			*s = runeValue('\n')
		} else if val[0] == '\\' && val[1] == 'r' {
			*s = runeValue('\r')
		} else if val[0] == '\\' && val[1] == 't' {
			*s = runeValue('\t')
		} else {
			return errors.New("Only \\n, \\r, and \\t are accepted escaped characters")
		}
	} else {
		return errors.New("Must contain a single character or escaped character")
	}

	return nil
}

func (s *runeValue) Get() interface{} { return rune(*s) }

func (s *runeValue) String() string { return fmt.Sprintf("%s", *s) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }

// -- stringMap Value
type stringMapValue map[string]string

func newStringMapValue(val map[string]string, p *map[string]string) *stringMapValue {
	*p = val
	return (*stringMapValue)(p)
}

func (s *stringMapValue) Set(val string) error {
	mappingsList := strings.Split(val, ",")
	for _, mapping := range mappingsList {
		pair := strings.Split(mapping, "=")
		if len(pair) != 2 {
			return fmt.Errorf("Mapping `%s` should contain two parts", mapping)
		}

		if pair[0] == "" || pair[1] == "" {
			return fmt.Errorf("Empty string in bucket mapping `%s`", mapping)
		}

		(map[string]string)(*s)[pair[0]] = pair[1]
	}

	return nil
}

func (s *stringMapValue) Get() interface{} { return map[string]string(*s) }

func (s *stringMapValue) String() string { return fmt.Sprintf("%s", *s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return fmt.Sprintf("%v", *f) }

// -- int Array
type intArray []int

func newIntArray(val []int, p *[]int) *intArray {
	*p = val
	return (*intArray)(p)
}

func (i *intArray) Set(s string) error {
	match := regexp.MustCompile(`(\d+,)*(\d+)`).FindString(s)
	if len(match) != len(s) {
		return errors.New("Not a list of integers")
	}

	elems := strings.Split(s, ",")
	for _, elem := range elems {
		val, err := strconv.Atoi(elem)
		if err != nil {
			return err
		}
		*i = append(*i, val)
	}

	return nil
}

func (i *intArray) Get() interface{} { return []int(*i) }

func (i *intArray) String() string { return fmt.Sprintf("%v", *i) }

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
type Value interface {
	String() string
	Set(string) error
}

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface.
type Getter interface {
	Value
	Get() interface{}
}

package goNixArgParser

import (
	"bytes"
	"errors"
	"os"
	"strings"
)

func StringToSlice(input string) []string {
	if len(input) == 0 {
		return nil
	}

	return []string{input}
}

func NewOptionSet(
	mergeFlagPrefix string,
	restsSigns []string,
	groupSeps []string,
	undefFlagPrefixes []string,
) *OptionSet {
	s := &OptionSet{
		mergeFlagPrefix:   mergeFlagPrefix,
		restsSigns:        restsSigns,
		groupSeps:         groupSeps,
		undefFlagPrefixes: undefFlagPrefixes,

		options: []*Option{},

		keyOptionMap:  map[string]*Option{},
		flagOptionMap: map[string]*Option{},
		flagMap:       map[string]*Flag{},
		keyEnvMap:     map[string][]string{},
		keyDefaultMap: map[string][]string{},
	}
	return s
}

func (s *OptionSet) MergeFlagPrefix() string {
	return s.mergeFlagPrefix
}

func (s *OptionSet) RestsSigns() []string {
	return s.restsSigns
}

func (s *OptionSet) GroupSeps() []string {
	return s.groupSeps
}

func (s *OptionSet) UndefFlagPrefixes() []string {
	return s.undefFlagPrefixes
}

func NewSimpleOptionSet() *OptionSet {
	return NewOptionSet("-", []string{"--"}, []string{",,"}, []string{"-"})
}

func (s *OptionSet) isRestSign(input string) bool {
	for _, sign := range s.restsSigns {
		if input == sign {
			return true
		}
	}

	return false
}

func (s *OptionSet) isGroupSep(input string) bool {
	for _, sep := range s.groupSeps {
		if input == sep {
			return true
		}
	}

	return false
}

func (s *OptionSet) isUdefFlag(input string) bool {
	for _, prefix := range s.undefFlagPrefixes {
		if strings.HasPrefix(input, prefix) {
			return true
		}
	}

	return false
}

func (s *OptionSet) Append(opt Option) error {
	// verify
	if len(opt.Key) == 0 {
		return errors.New("key is empty")
	}
	if s.keyOptionMap[opt.Key] != nil {
		return errors.New("key '" + opt.Key + "' already exists")
	}

	for _, flag := range opt.Flags {
		flagName := flag.Name
		if len(flagName) == 0 {
			return errors.New("flag name is empty")
		}
		if s.flagMap[flagName] != nil {
			return errors.New("flag '" + flagName + "' already exists")
		}
	}

	option := &opt

	// append
	s.options = append(s.options, option)

	// redundant - flag summaries, maps
	s.keyOptionMap[option.Key] = option
	for _, flag := range option.Flags {
		if flag.canMerge {
			s.hasCanMerge = true
		}
		if flag.canConcatAssign {
			s.hasCanConcatAssign = true
		}
		if len(flag.assignSigns) > 0 {
			s.hasAssignSigns = true
		}

		flagName := flag.Name
		s.flagOptionMap[flagName] = option
		s.flagMap[flagName] = flag
	}

	// redundant - env maps
	if len(option.EnvVars) > 0 {
		for _, envVar := range option.EnvVars {
			if len(envVar) == 0 {
				continue
			}
			envValue, hasEnv := os.LookupEnv(envVar)
			if !hasEnv || len(envValue) == 0 {
				continue
			}

			if option.MultiValues {
				s.keyEnvMap[option.Key] = option.splitValues(envValue)
			} else {
				s.keyEnvMap[option.Key] = []string{envValue}
			}
		}
	}

	// redundant - default maps
	if len(option.DefaultValues) > 0 {
		s.keyDefaultMap[option.Key] = option.DefaultValues
	}
	return nil
}

func (s *OptionSet) AddFlag(key, flag, envVar, summary string) error {
	return s.Append(Option{
		Key:     key,
		Flags:   []*Flag{NewSimpleFlag(flag)},
		EnvVars: StringToSlice(envVar),
		Summary: summary,
	})
}

func (s *OptionSet) AddFlags(key string, flags []string, envVar, summary string) error {
	return s.Append(Option{
		Key:     key,
		Flags:   NewSimpleFlags(flags),
		EnvVars: StringToSlice(envVar),
		Summary: summary,
	})
}

func (s *OptionSet) AddFlagValue(key, flag, envVar, defaultValue, summary string) error {
	return s.Append(Option{
		Key:           key,
		Flags:         []*Flag{NewSimpleFlag(flag)},
		AcceptValue:   true,
		OverridePrev:  true,
		EnvVars:       StringToSlice(envVar),
		DefaultValues: StringToSlice(defaultValue),
		Summary:       summary,
	})
}

func (s *OptionSet) AddFlagValues(key, flag, envVar string, defaultValues []string, summary string) error {
	return s.Append(Option{
		Key:           key,
		Flags:         []*Flag{NewSimpleFlag(flag)},
		AcceptValue:   true,
		MultiValues:   true,
		UniqueValues:  true,
		EnvVars:       StringToSlice(envVar),
		DefaultValues: defaultValues,
		Summary:       summary,
	})
}

func (s *OptionSet) AddFlagsValue(key string, flags []string, envVar, defaultValue, summary string) error {
	return s.Append(Option{
		Key:           key,
		Flags:         NewSimpleFlags(flags),
		AcceptValue:   true,
		OverridePrev:  true,
		EnvVars:       StringToSlice(envVar),
		DefaultValues: StringToSlice(defaultValue),
		Summary:       summary,
	})
}

func (s *OptionSet) AddFlagsValues(key string, flags []string, envVar string, defaultValues []string, summary string) error {
	return s.Append(Option{
		Key:           key,
		Flags:         NewSimpleFlags(flags),
		AcceptValue:   true,
		MultiValues:   true,
		UniqueValues:  true,
		EnvVars:       StringToSlice(envVar),
		DefaultValues: defaultValues,
		Summary:       summary,
	})
}

func (s *OptionSet) GetHelp() []byte {
	buffer := &bytes.Buffer{}
	for _, opt := range s.options {
		buffer.Write(opt.GetHelp())
		buffer.WriteByte('\n')
	}

	return buffer.Bytes()
}

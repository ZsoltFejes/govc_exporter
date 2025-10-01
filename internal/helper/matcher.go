package helper

import (
	"regexp"
	"strings"
)

type Matchable interface {
	Match(string) bool
}

type Matcher struct {
	Keywords []string
}

func (m *Matcher) First() string {
	if len(m.Keywords) > 0 {
		return m.Keywords[0]
	}
	return "none"
}

func NewMatcher(keywords ...string) *Matcher {
	return &Matcher{
		Keywords: keywords,
	}
}

func (m Matcher) Match(s string) bool {
	for _, keyword := range m.Keywords {
		if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
			return true
		}
	}
	return false
}

// MatchRegex matches the input string against all keywords treated as regex patterns.
// It returns true if any of the patterns match, otherwise false.
// If any regex pattern is invalid, it returns an error.
func (m Matcher) MatchRegex(s string) (bool, error) {
	for _, keyword := range m.Keywords {
		// treat wildcard/all as always matching without compiling regex
		if strings.ToLower(keyword) == "all" || keyword == "*" {
			return true, nil
		}
		re, err := regexp.Compile(keyword)
		if err != nil {
			return false, err
		}
		if re.MatchString(s) {
			return true, nil
		}
	}
	return false, nil
}

func (m Matcher) MatchAny(s ...string) bool {
	for _, keyword := range m.Keywords {
		for _, s := range s {
			if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
				return true
			}
		}
	}
	return false
}

func (m Matcher) MatchAll(s ...string) bool {
	for _, keyword := range m.Keywords {
		match := true
		for _, s := range s {
			if !(strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*") {
				match = false
			}
		}
		if match {
			return true
		}
	}
	return false
}

func AnyMatch(o Matchable, keywords ...string) bool {
	for _, keyword := range keywords {
		match := o.Match(keyword)
		if match {
			return true
		}
	}
	return false
}

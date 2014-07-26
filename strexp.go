package strexp

import (
	"regexp"
)

type StrExpConfig struct {
	requirements map[string]string
	separators   string
}

func (self *StrExpConfig) hasPattern(name string) bool {
	_, r := self.requirements[name]

	return r
}

func (self *StrExpConfig) patternForName(name string) string {
	if p, ok := self.requirements[name]; ok == true {
		return p
	}

	if len(self.separators) > 0 {
		return "[^" + self.separators + "]+"
	}

	return ".+"
}

// Compile parses a segmented string expression and compiles it into a Regexp
func Compile(strexp string, c *StrExpConfig) (*regexp.Regexp, error) {
	t, err := Parse(strexp)
	if err != nil {
		return nil, err
	}

	r := t.RegExpFragment(c)

	return regexp.Compile("^" + r + "$")
}

func MustCompile(strexp string, c *StrExpConfig) *regexp.Regexp {
	r, err := Compile(strexp, c)
	if err != nil {
		panic(err)
	}

	return r
}

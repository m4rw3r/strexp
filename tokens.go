package strexp

import (
	"regexp"
)

type Token interface {
	RegExpFragment(*StrExpConfig) string
	String() string
}

type tList struct {
	items []Token
}

func (self tList) RegExpFragment(c *StrExpConfig) string {
	str := ""

	for _, i := range self.items {
		str += i.RegExpFragment(c)
	}

	return str
}

func (self tList) String() string {
	str := ""

	for _, i := range self.items {
		str += i.String()
	}

	return str
}

type tGroup struct {
	item Token
}

func (self tGroup) RegExpFragment(c *StrExpConfig) string {
	return "(?:" + self.item.RegExpFragment(c) + ")?"
}

func (self tGroup) String() string {
	return "(" + self.item.String() + ")"
}

type tParam struct {
	name string
}

func (self tParam) RegExpFragment(c *StrExpConfig) string {
	pattern := c.patternForName(self.name)

	return "(?<" + self.name + ">" + pattern + ")"
}

func (self tParam) String() string {
	return ":" + self.name
}

type tGlob struct {
	name string
}

func (self tGlob) RegExpFragment(c *StrExpConfig) string {
	pattern := c.patternForName(self.name)

	if !c.hasPattern(self.name) {
		pattern = ".+"
	}

	return "(?<" + self.name + ">" + pattern + ")"
}

func (self tGlob) String() string {
	return "*" + self.name
}

type tChar struct {
	char string
}

func (self tChar) RegExpFragment(c *StrExpConfig) string {
	return regexp.QuoteMeta(self.char)
}

func (self tChar) String() string {
	reserved := []byte{
		40, // (
		41, // )
		42, // *
		58, // :
		92, // \
	}

	for _, c := range reserved {
		if c == self.char[0] {
			return "\\" + self.char
		}
	}

	return self.char
}

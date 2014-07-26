package strexp

import (
	"errors"
	"strings"
)

// All parsing functions have a prototype similar to this:
// func(str string) (result, string, error)
// str:    The full remainder of the string to parse
// result: The expected result, if successful parsing
// string: The remainder of the string to parse, after result has been parsed
// error:  If any parse error has occured, this is non-nil

// pName := [a-zA-Z_]+
func pName(str string) (string, string, error) {
	name := ""

	for {
		if len(str) == 0 {
			break
		}

		if str[0] > 65 && str[0] < 91 || str[0] > 96 && str[0] < 123 || str[0] == 95 {
			name += str[0:1]
			str = str[1:]
		} else {
			break
		}
	}

	if len(name) == 0 {
		return "", str, errors.New("unexpected character: '" + str[0:1] + "', expected identifier.")
	}

	return name, str, nil
}

// pExpect attempts to match the beginning of str with char, if it succeeds
// it will return str with the char-prefix removed and nil,
// returns an error if it does not match
func pExpect(str string, char string) (string, error) {
	if !strings.HasPrefix(str, char) {
		if len(str) < len(char) {
			return str, errors.New("unexpected end, expected '" + char + "'.")
		}

		return str, errors.New("unexpected character: '" + str[0:len(char)] + "', expected '" + char + "'.")
	}

	return str[len(char):], nil
}

// pParam := ':' pName
func pParam(str string) (Token, string, error) {
	str, err := pExpect(str, ":")
	if err != nil {
		return tList{}, str, err
	}

	name, str, err := pName(str)
	if err != nil {
		return tList{}, str, err
	}

	return tParam{name}, str, nil
}

// pGlob := '*' pName
func pGlob(str string) (Token, string, error) {
	str, err := pExpect(str, "*")
	if err != nil {
		return tList{}, str, err
	}

	name, str, err := pName(str)
	if err != nil {
		return tList{}, str, err
	}

	return tGlob{name}, str, nil
}

// pGroup := '(' pExprGroup ')'
func pGroup(str string) (Token, string, error) {
	s, err := pExpect(str, "(")
	if err != nil {
		return tList{}, s, err
	}

	t, s, err := pExprGroup(s)
	if err != nil {
		return tList{}, s, err
	}

	s, err = pExpect(s, ")")
	if err != nil {
		return tList{}, s, err
	}

	return tGroup{t}, s, nil
}

// pChar := '\' ('(' | ')' | '*' | ':' | '\') | [^()*:\]
func pChar(str string) (Token, string, error) {
	reserved := []byte{
		40, // (
		41, // )
		42, // *
		58, // :
		92, // \
	}

	if str[0] == 92 {
		/* Escape char */
		for _, c := range reserved {
			if str[1] == c {
				return tChar{str[1:1]}, str[2:], nil
			}
		}

		return tList{}, str, errors.New("Invalid escape-sequence '\\" + str[1:1] + "', expected one of '\\(', '\\)', '\\*', '\\:' or '\\\\'.")
	}

	/* Make sure we do not use these, syntax error */
	for _, c := range reserved {
		if str[0] == c {
			return tList{}, str, errors.New("Unexpected special character '" + str[0:1] + "', missing escaping.")
		}
	}

	return tChar{str[0:1]}, str[1:], nil
}

// pToken := pParam | pGlob | pGroup | pChar
func pToken(str string) (Token, string, error) {
	/* Peek */
	_, err := pExpect(str, ":")
	if err == nil {
		/* Explicitly peek for special character to produce proper error messages,
		   only a param starts with ':' */
		t, s, err := pParam(str)
		if err == nil {
			return t, s, nil
		}

		return tList{}, s, err
	}

	_, err = pExpect(str, "*")
	if err == nil {
		t, s, err := pGlob(str)
		if err == nil {
			return t, s, nil
		}

		return tList{}, s, err
	}

	_, err = pExpect(str, "(")
	if err == nil {
		t, s, err := pGroup(str)
		if err == nil {
			return t, s, nil
		}

		return tList{}, s, err
	}

	t, s, err := pChar(str)
	if err == nil {
		return t, s, nil
	}

	return tList{}, str, err
}

// pExprGroup := pToken+ EOF | pToken+ ')'
// The reason why it also ends on ')' is to allow for pGroup to continue
// parsing and avoid failure on ')'.
// Parse checks for premature exit
func pExprGroup(str string) (Token, string, error) {
	var s string = str
	var t Token
	var err error

	tokens := make([]Token, 0)

	for {
		if len(s) == 0 {
			break
		}

		if _, err = pExpect(s, ")"); err == nil {
			break
		}

		t, s, err = pToken(s)
		if err != nil {
			return tList{}, "", err
		}

		tokens = append(tokens, t)
	}

	return tList{tokens}, s, nil
}

// Parse parses a strexp string into a Token instance
func Parse(str string) (Token, error) {
	t, s, err := pExprGroup(str)

	if err == nil && len(s) != 0 {
		/* Only way for it to break with a shorter string is for it to encounter an unbalanced ")" */

		err = errors.New("Unmatched ')'.")
	}

	if err != nil {
		return tList{}, errors.New("strexp parse error: after '" + str[0:len(str)-len(s)] + "': " + err.Error())
	}

	return t, nil
}

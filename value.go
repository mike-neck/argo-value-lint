package main

import (
	"fmt"
	"strings"
)

type LintConfig struct {
	Sprig
}

type Value []Token

type Token struct {
	Type TokenType
	Text string
}

func (t Token) String() string {
	return fmt.Sprintf("%s[%s]", t.Type, t.Text)
}

type TokenType int

const (
	RawValue TokenType = iota
	Expression
	Variable
)

func (t TokenType) String() string {
	switch t {
	case RawValue:
		return "RawValue"
	case Expression:
		return "Expression"
	case Variable:
		return "Variable"
	}
	panic(fmt.Sprintf("unknown token type: %d", t))
}

type TokenValidator interface {
	Validate(text string) error
}

func ParseValue(value string) Value {
	vs := make(Value, 0)
	exp := false
	var sb strings.Builder
	runes := []rune(value)
	length := len(runes)
	for i := 0; i < length; i++ {
		switch runes[i] {
		case '{':
			if exp {
				sb.WriteRune('{')
			} else if i == length-1 {
				vs = append(vs, Token{Type: RawValue, Text: sb.String()})
				sb.Reset()
			} else if runes[i+1] == '{' {
				exp = true
				if 0 < sb.Len() {
					vs = append(vs, Token{Type: RawValue, Text: sb.String()})
					sb.Reset()
				}
				i++
			} else {
				sb.WriteRune('{')
			}
		case '}':
			if exp && i == length-1 {
				if len(vs) == 0 {
					t := fmt.Sprintf("{{%s}", sb.String())
					vs = append(vs, Token{Type: RawValue, Text: t})
				} else {
					preToken := vs[len(vs)-1]
					nextText := fmt.Sprintf("{{%s%s}", preToken.Text, sb.String())
					preToken.Type = RawValue
					preToken.Text = nextText
				}
				sb.Reset()
			} else if exp {
				if runes[i+1] == '}' {
					exp = false
					expression := strings.TrimSpace(sb.String())
					if strings.HasPrefix(expression, "=") {
						vs = append(vs, Token{Type: Expression, Text: expression[1:]})
					} else {
						vs = append(vs, Token{Type: Variable, Text: expression})
					}
					sb.Reset()
					i++
				} else {
					sb.WriteRune('}')
				}
			} else {
				sb.WriteRune('}')
			}
		case '\\':
			if i == length-1 {
				// skip this rune
			} else {
				switch runes[i+1] {
				case '\\':
					sb.WriteRune('\\')
				case '"':
					sb.WriteRune('"')
				case '\'':
					sb.WriteRune('\'')
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case 'f':
					sb.WriteRune('\f')
				case '{':
					sb.WriteRune('{')
				case '}':
					sb.WriteRune('}')
				}
				i++
			}
		default:
			sb.WriteRune(runes[i])
		}
	}
	if 0 < sb.Len() {
		t := sb.String()
		if exp {
			if 0 < len(vs) {
				preToken := vs[len(vs)-1]
				t = fmt.Sprintf("{{%s%s", preToken.Text, t)
			} else {
				t = fmt.Sprintf("{{%s", t)
			}
			token := Token{Type: RawValue, Text: t}
			if 0 < len(vs) {
				vs[len(vs)-1] = token
			} else {
				vs = append(vs, token)
			}
		} else {
			vs = append(vs, Token{Type: RawValue, Text: t})
		}
	}
	return vs
}

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

type TokenType int

const (
	RawValue TokenType = iota
	Expression
)

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
					vs = append(vs, Token{Type: Expression, Text: sb.String()})
					sb.Reset()
					i++
				} else {
					sb.WriteRune('}')
				}
			} else {
				sb.WriteRune('}')
			}
		default:
			sb.WriteRune(runes[i])
		}
	}
	if 0 < sb.Len() {
		t := sb.String()
		if exp {
			preToken := vs[len(vs)-1]
			nextText := fmt.Sprintf("{{%s%s", preToken.Text, t)
			preToken.Type = RawValue
			preToken.Text = nextText
		} else {
			vs = append(vs, Token{Type: RawValue, Text: t})
		}
	}
	return vs
}

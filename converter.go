package gof

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type Converter struct {
	in  *bufio.Reader
	out *bufio.Writer
}

type stateFunc func() stateFunc

func Convert(i io.Reader, o io.Writer) {
	c := &Converter{in: bufio.NewReader(i), out: bufio.NewWriter(o)}
	c.start()
}

func (c *Converter) start() {
	nextFunc := c.processCode
	for nextFunc != nil {
		nextFunc = nextFunc()
	}
	if err := c.out.Flush(); err != nil {
		panic(err)
	}
}

func (c *Converter) processCode() stateFunc {
	var lastRune rune
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		resetLastRune := func() {
			lastRune = 0
			_, err := c.out.WriteRune(lastRune)
			if err != nil {
				panic(err)
			}
			_, err = c.out.WriteRune(r)
			if err != nil {
				panic(err)
			}
		}
		switch r {
		case '"':
			if lastRune == '$' {
				return c.processFormatString
			}
			return c.processStringLiteral
		case '`':
			return c.processMultilineStringLiteral
		case '\'':
			return c.processRuneLiteral
		case '/':
			if lastRune == 0 {
				lastRune = r
				break
			}
			if lastRune == '/' {
				return c.processComment
			}
			resetLastRune()
		case '*':
			if lastRune == 0 {
				_, err = c.out.WriteRune(r)
				if err != nil {
					panic(err)
				}
				lastRune = 0
				break
			}
			if lastRune == '/' {
				return c.processMultilineComment
			}
			resetLastRune()
		case '$':
			lastRune = r
		default:
			lastRune = 0
			_, err = c.out.WriteRune(r)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func (c *Converter) processStringLiteral() stateFunc {
	_, err := c.out.WriteRune('"')
	if err != nil {
		panic(err)
	}
	var escaped bool
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		if r == '\\' {
			if escaped {
				_, err := c.out.WriteString("\\\\")
				if err != nil {
					panic(err)
				}
				escaped = false
				continue
			}
			escaped = true
			continue
		}
		if r == '"' {
			if escaped {
				_, err := c.out.WriteString("\\\"")
				if err != nil {
					panic(err)
				}
				continue
			}
			_, err = c.out.WriteString("\"")
			if err != nil {
				panic(err)
			}
			return c.processCode
		}
		if escaped {
			_, err := c.out.WriteRune('\\')
			if err != nil {
				panic(err)
			}
			escaped = false
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (c *Converter) processMultilineStringLiteral() stateFunc {
	_, err := c.out.WriteRune('`')
	if err != nil {
		panic(err)
	}
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			panic(err)
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
		if r == '`' {
			return c.processCode
		}
	}
}

func (c *Converter) processRuneLiteral() stateFunc {
	_, err := c.out.WriteRune('\'')
	if err != nil {
		panic(err)
	}
	var escaped bool
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			panic(err)
		}
		if r == '\\' {
			if escaped {
				_, err = c.out.WriteString("\\\\")
				if err != nil {
					panic(err)
				}
				escaped = false
				continue
			}
			escaped = true
			continue
		}
		if r == '\'' {
			if escaped {
				_, err = c.out.WriteString("\\'")
				if err != nil {
					panic(err)
				}
				escaped = false
				continue
			}
			_, err = c.out.WriteRune('\'')
			if err != nil {
				panic(err)
			}
			return c.processCode
		}
		if escaped {
			_, err = c.out.WriteRune('\\')
			if err != nil {
				panic(err)
			}
			escaped = false
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Converter) processComment() stateFunc {
	_, err := c.out.WriteString("//")
	if err != nil {
		panic(err)
	}
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			panic(err)
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
		if r == '\n' {
			return c.processCode
		}
	}
}

func (c *Converter) processMultilineComment() stateFunc {
	_, err := c.out.WriteString("/*")
	if err != nil {
		panic(err)
	}
	var lastRuneWasAsterisk bool
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			panic(err)
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
		if r == '/' && lastRuneWasAsterisk {
			return c.processCode
		}
		lastRuneWasAsterisk = r == '*'
	}
}

func (c *Converter) processFormatString() stateFunc {
	_, err := c.out.WriteRune('"')
	if err != nil {
		panic(err)
	}
	expressions := make([]string, 0)
	var escaped bool
outer:
	for {
		r, _, err := c.in.ReadRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			panic(err)
		}
		if r == '\\' {
			if escaped {
				_, err = c.out.WriteString("\\\\")
				if err != nil {
					panic(err)
				}
				escaped = false
				continue
			}
			escaped = true
			continue
		}
		if r == '"' {
			if escaped {
				_, err := c.out.WriteString("\\\"")
				if err != nil {
					panic(err)
				}
				continue
			}
			_, err = c.out.WriteRune('"')
			if err != nil {
				panic(err)
			}
			for _, expression := range expressions {
				c.out.WriteString(", " + expression)
			}
			return c.processCode
		}
		if r == '{' {
			if escaped {
				_, err := c.out.WriteRune('{')
				if err != nil {
					panic(err)
				}
				continue
			}
			expressionBuilder := new(strings.Builder)
			for {
				r, _, err = c.in.ReadRune()
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					panic(err)
				}
				if r == '}' {
					expressions = append(expressions, expressionBuilder.String())
					continue outer
				}
				expressionBuilder.WriteRune(r)
			}
		}
		if escaped {
			_, err = c.out.WriteRune('\\')
			if err != nil {
				panic(err)
			}
		}
		_, err = c.out.WriteRune(r)
		if err != nil {
			panic(err)
		}
	}
}

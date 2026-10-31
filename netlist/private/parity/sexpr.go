// SPDX-License-Identifier: Apache-2.0

package parity

import (
	"fmt"
	"strings"
)

// Node is an S-expression: a list (List set) or an atom (Atom set; a
// quoted string is an atom too, unquoted).
type Node struct {
	Atom   string
	List   []*Node
	IsList bool
}

// Head is a list's first atom: "net" for (net 1 "GND").
func (n *Node) Head() string {
	if !n.IsList || len(n.List) == 0 || n.List[0].IsList {
		return ""
	}
	return n.List[0].Atom
}

// Child returns the first sub-list with this head, or nil.
func (n *Node) Child(head string) *Node {
	for _, c := range n.List {
		if c.Head() == head {
			return c
		}
	}
	return nil
}

// Children returns every sub-list with this head.
func (n *Node) Children(head string) []*Node {
	var out []*Node
	for _, c := range n.List {
		if c.Head() == head {
			out = append(out, c)
		}
	}
	return out
}

// Arg returns the i-th atom after the head ("" if there is none).
func (n *Node) Arg(i int) string {
	k := 0
	for _, c := range n.List[1:] {
		if !c.IsList {
			if k == i {
				return c.Atom
			}
			k++
		}
	}
	return ""
}

// Parse reads one S-expression, as KiCad writes them.
func Parse(src string) (*Node, error) {
	p := &parser{s: src}
	p.space()
	n, err := p.node()
	if err != nil {
		return nil, err
	}
	p.space()
	if p.i != len(p.s) {
		return nil, fmt.Errorf("offset %d: text after the expression", p.i)
	}
	return n, nil
}

type parser struct {
	s string
	i int
}

// The delimiters are ASCII, so the parser works on bytes; a multi-byte
// UTF-8 character never holds one of them.
func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\r' || b == '\n' }

func (p *parser) space() {
	for p.i < len(p.s) && isSpace(p.s[p.i]) {
		p.i++
	}
}

func (p *parser) node() (*Node, error) {
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("offset %d: unexpected end", p.i)
	}
	switch p.s[p.i] {
	case '(':
		p.i++
		n := &Node{IsList: true}
		for {
			p.space()
			if p.i >= len(p.s) {
				return nil, fmt.Errorf("offset %d: unclosed list", p.i)
			}
			if p.s[p.i] == ')' {
				p.i++
				return n, nil
			}
			c, err := p.node()
			if err != nil {
				return nil, err
			}
			n.List = append(n.List, c)
		}
	case ')':
		return nil, fmt.Errorf("offset %d: unexpected )", p.i)
	case '"':
		p.i++
		var b strings.Builder
		for p.i < len(p.s) && p.s[p.i] != '"' {
			c := p.s[p.i]
			if c == '\\' && p.i+1 < len(p.s) {
				p.i++
				// KiCad writes C escapes in strings.
				switch c = p.s[p.i]; c {
				case 'n':
					c = '\n'
				case 't':
					c = '\t'
				case 'r':
					c = '\r'
				}
			}
			b.WriteByte(c)
			p.i++
		}
		if p.i >= len(p.s) {
			return nil, fmt.Errorf("offset %d: unclosed string", p.i)
		}
		p.i++
		return &Node{Atom: b.String()}, nil
	default:
		j := p.i
		for p.i < len(p.s) && !isSpace(p.s[p.i]) && p.s[p.i] != '(' && p.s[p.i] != ')' && p.s[p.i] != '"' {
			p.i++
		}
		return &Node{Atom: p.s[j:p.i]}, nil
	}
}

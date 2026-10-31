// SPDX-License-Identifier: Apache-2.0

// Package parity compares the nets of a KiCad netlist with those of a
// board (issue #11): which pins each net joins. Net names are not
// compared, so a net SKiDL names N$1 and KiCad's schematic Net-(D1-A)
// match when they join the same pins.
package parity

import (
	"fmt"
	"sort"
	"strings"
)

// Pin is a component's pin, "R1.2".
type Pin = string

// Nets maps each net's name to its pins, each pin once, in file order.
type Nets map[string][]Pin

// Design is what the comparison reads from a netlist or a board.
type Design struct {
	Refs map[string]bool // every component reference
	Nets Nets
	// Conflicts are pins on two nets at once: a footprint with two pads
	// of one number (an SOT-223's tab, a QFN's ground pads) on different
	// nets. Compare reports them.
	Conflicts []string
}

// ReadNetlist reads a KiCad netlist (kicad_netlist's or skidl_netlist's).
func ReadNetlist(src string) (*Design, error) {
	root, err := Parse(src)
	if err != nil {
		return nil, fmt.Errorf("netlist: %v", err)
	}
	if root.Head() != "export" {
		return nil, fmt.Errorf("netlist: not a KiCad netlist (it starts with %q)", root.Head())
	}
	d := &Design{Refs: map[string]bool{}, Nets: Nets{}}
	if c := root.Child("components"); c != nil {
		for _, comp := range c.Children("comp") {
			if r := comp.Child("ref"); r != nil {
				d.Refs[r.Arg(0)] = true
			}
		}
	}
	nets := root.Child("nets")
	if nets == nil {
		return nil, fmt.Errorf("netlist: no (nets ...)")
	}
	for _, n := range nets.Children("net") {
		name := ""
		if c := n.Child("name"); c != nil {
			name = c.Arg(0)
		}
		for _, node := range n.Children("node") {
			r, p := node.Child("ref"), node.Child("pin")
			if r == nil || p == nil {
				return nil, fmt.Errorf("netlist: net %q has a node without ref or pin", name)
			}
			pin := r.Arg(0) + "." + p.Arg(0)
			if !contains(d.Nets[name], pin) {
				d.Nets[name] = append(d.Nets[name], pin)
			}
		}
	}
	return d, nil
}

// ReadBoard reads a .kicad_pcb: each footprint's reference, and the net of
// each of its pads. A pad on no net (net 0, or no net at all) is left out.
func ReadBoard(src string) (*Design, error) {
	root, err := Parse(src)
	if err != nil {
		return nil, fmt.Errorf("board: %v", err)
	}
	if root.Head() != "kicad_pcb" {
		return nil, fmt.Errorf("board: not a .kicad_pcb (it starts with %q)", root.Head())
	}
	d := &Design{Refs: map[string]bool{}, Nets: Nets{}}
	netOf := map[Pin]string{}
	for _, fp := range root.Children("footprint") {
		ref := ""
		for _, prop := range fp.Children("property") {
			if prop.Arg(0) == "Reference" {
				ref = prop.Arg(1)
			}
		}
		if ref == "" {
			// KiCad 7 and before: (fp_text reference "R1" ...).
			for _, t := range fp.Children("fp_text") {
				if t.Arg(0) == "reference" {
					ref = t.Arg(1)
				}
			}
		}
		if ref == "" {
			return nil, fmt.Errorf("board: a footprint %q has no reference", fp.Arg(0))
		}
		d.Refs[ref] = true
		for _, pad := range fp.Children("pad") {
			n := pad.Child("net")
			if n == nil {
				continue
			}
			// (net 2 "GND") in KiCad 9; a name alone in later versions;
			// (net 0 "") or (net 0) for none.
			a := atoms(n)
			if len(a) == 0 {
				continue
			}
			name := a[len(a)-1]
			if name == "" || len(a) == 1 && name == "0" {
				continue
			}
			pin := ref + "." + pad.Arg(0)
			if prev, ok := netOf[pin]; ok && prev != name {
				d.Conflicts = append(d.Conflicts, fmt.Sprintf("pin %s: pads on board nets %s and %s", pin, prev, name))
				continue
			}
			netOf[pin] = name
			if !contains(d.Nets[name], pin) {
				d.Nets[name] = append(d.Nets[name], pin)
			}
		}
	}
	return d, nil
}

func atoms(n *Node) []string {
	var out []string
	for _, c := range n.List[1:] {
		if !c.IsList {
			out = append(out, c.Atom)
		}
	}
	return out
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// pinNet maps each pin to its net, for nets of at least min pins.
func pinNet(n Nets, min int) map[Pin]string {
	m := map[Pin]string{}
	for name, pins := range n {
		if len(pins) < min {
			continue
		}
		for _, p := range pins {
			m[p] = name
		}
	}
	return m
}

// Compare returns each difference between the netlist and the board,
// sorted; none when the board joins exactly the pins the netlist joins.
func Compare(netlist, board *Design) []string {
	var problems []string
	add := func(f string, a ...any) { problems = append(problems, fmt.Sprintf(f, a...)) }
	problems = append(problems, board.Conflicts...)
	for ref := range netlist.Refs {
		if !board.Refs[ref] {
			add("component %s: in the netlist, not on the board", ref)
		}
	}
	// A netlist net of one pin connects nothing (KiCad lists unconnected
	// pins so); on the board, a pin on a net of its own is still on it.
	nl, bd := pinNet(netlist.Nets, 2), pinNet(board.Nets, 1)
	for _, name := range sortedKeys(netlist.Nets) {
		pins := netlist.Nets[name]
		if len(pins) < 2 {
			continue
		}
		on := map[string][]Pin{}
		for _, p := range sorted(pins) {
			b, ok := bd[p]
			if !ok {
				add("net %s: pin %s is on no net of the board", name, p)
				continue
			}
			on[b] = append(on[b], p)
		}
		if len(on) > 1 {
			var parts []string
			for _, b := range sortedKeys(on) {
				parts = append(parts, fmt.Sprintf("%s (%s)", b, strings.Join(on[b], " ")))
			}
			add("net %s: split on the board into %s", name, strings.Join(parts, ", "))
		}
	}
	for _, name := range sortedKeys(board.Nets) {
		pins := board.Nets[name]
		if len(pins) < 2 {
			continue
		}
		in := map[string]bool{}
		for _, p := range sorted(pins) {
			n, ok := nl[p]
			if !ok {
				add("board net %s: pin %s is on no net of the netlist", name, p)
				continue
			}
			in[n] = true
		}
		if len(in) > 1 {
			add("board net %s: joins netlist nets %s", name, strings.Join(sortedKeys(in), ", "))
		}
	}
	sort.Strings(problems)
	return problems
}

func sorted(l []string) []string {
	c := append([]string(nil), l...)
	sort.Strings(c)
	return c
}

func sortedKeys[V any](m map[string]V) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}

// Verdict is the test's result: with want empty, an error if there are
// problems; with want set (a known-bad board), an error unless one of
// the problems contains it.
func Verdict(problems []string, want string) error {
	if want == "" {
		if len(problems) > 0 {
			return fmt.Errorf("%d differences", len(problems))
		}
		return nil
	}
	for _, p := range problems {
		if strings.Contains(p, want) {
			return nil
		}
	}
	return fmt.Errorf("want a difference containing %q; got %d others", want, len(problems))
}

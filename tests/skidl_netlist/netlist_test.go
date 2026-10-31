// SPDX-License-Identifier: Apache-2.0

// Package netlist_test reads the netlist skidl_netlist writes for the LED
// circuit: its fixed header, its parts and its nets.
package netlist_test

import (
	"flag"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

var netlist = flag.String("netlist", "", "the netlist, as a runfiles path")

func read(t *testing.T) string {
	t.Helper()
	p, err := runfiles.Rlocation(*netlist)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestHeader(t *testing.T) {
	n := read(t)
	for _, want := range []string{
		`(source "tests/skidl_netlist/circuit.py")`,
		`(date "1970-01-01T00:00:00+0000")`,
		`(tool "SKiDL (3.0.0)")`,
	} {
		if !strings.Contains(n, want) {
			t.Errorf("no %s in the header", want)
		}
	}
	if s := regexp.MustCompile(`\(source "([^"]*)"\)`).FindAllStringSubmatch(n, -1); len(s) == 0 {
		t.Error("no (source ...)")
	} else {
		for _, m := range s {
			if m[1] != "tests/skidl_netlist/circuit.py" {
				t.Errorf("(source %q): a path that depends on where the action ran", m[1])
			}
		}
	}
}

var (
	compRe = regexp.MustCompile(`\(comp\s+\(ref "([^"]+)"\)\s+\(value "([^"]*)"\)\s+(?:\(description "[^"]*"\)\s+)?\(footprint "([^"]*)"\)`)
	lineRe = regexp.MustCompile(`\(name "SKiDL Line"\) "([^"]*)"`)
	netRe  = regexp.MustCompile(`\(net\s+\(code \d+\)\s+\(name "([^"]*)"\)`)
	nodeRe = regexp.MustCompile(`\(node\s+\(ref "([^"]+)"\)\s+\(pin "([^"]+)"\)`)
)

func TestParts(t *testing.T) {
	n := read(t)
	got := map[string]string{}
	for _, m := range compRe.FindAllStringSubmatch(n, -1) {
		got[m[1]] = m[2] + " " + m[3]
	}
	want := map[string]string{
		"J1": "Power Connector_PinHeader_2.54mm:PinHeader_1x02_P2.54mm_Vertical",
		"R1": "1k Resistor_SMD:R_0603_1608Metric",
		"D1": "LED LED_SMD:LED_0603_1608Metric",
	}
	if len(got) != len(want) {
		t.Errorf("parts %v, want %v", got, want)
	}
	for ref, w := range want {
		if got[ref] != w {
			t.Errorf("%s: %q, want %q", ref, got[ref], w)
		}
	}
}

// Each part's SKiDL Line names its file by its path in the workspace, not
// by one that climbs out of the sandbox.
func TestSourceLines(t *testing.T) {
	got := lineRe.FindAllStringSubmatch(read(t), -1)
	if len(got) != 3 {
		t.Fatalf("%d SKiDL Line fields, want 3", len(got))
	}
	for _, m := range got {
		if !strings.HasPrefix(m[1], "tests/skidl_netlist/") {
			t.Errorf("SKiDL Line %q", m[1])
		}
	}
}

// TestNets checks the connections: each net's pins, by name for the named
// nets and by its pins for the one SKiDL names.
func TestNets(t *testing.T) {
	n := read(t)
	i := strings.Index(n, "(nets")
	if i < 0 {
		t.Fatal("no (nets ...)")
	}
	nets := map[string]string{}
	locs := netRe.FindAllStringSubmatchIndex(n[i:], -1)
	for k, l := range locs {
		end := len(n) - i
		if k+1 < len(locs) {
			end = locs[k+1][0]
		}
		var pins []string
		for _, m := range nodeRe.FindAllStringSubmatch(n[i+l[0]:i+end], -1) {
			pins = append(pins, m[1]+"."+m[2])
		}
		sort.Strings(pins)
		nets[n[i+l[2]:i+l[3]]] = strings.Join(pins, " ")
	}
	if nets["+3V3"] != "J1.1 R1.1" {
		t.Errorf("+3V3: %q", nets["+3V3"])
	}
	if nets["GND"] != "D1.1 J1.2" {
		t.Errorf("GND: %q", nets["GND"])
	}
	found := false
	for name, pins := range nets {
		if name != "+3V3" && name != "GND" {
			if pins != "D1.2 R1.2" {
				t.Errorf("%s: %q, want D1.2 R1.2", name, pins)
			}
			found = true
		}
	}
	if !found || len(nets) != 3 {
		t.Errorf("nets %v, want +3V3, GND and the LED's anode net", nets)
	}
}

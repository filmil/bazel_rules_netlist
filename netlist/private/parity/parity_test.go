// SPDX-License-Identifier: Apache-2.0

package parity

import (
	"strings"
	"testing"
)

const netlist = `(export (version "E")
  (components
    (comp (ref "J1") (value "Power"))
    (comp (ref "R1") (value "1k"))
    (comp (ref "D1") (value "LED")))
  (nets
    (net (code "1") (name "+3V3") (node (ref "J1") (pin "1")) (node (ref "R1") (pin "1")))
    (net (code "2") (name "GND") (node (ref "D1") (pin "1")) (node (ref "J1") (pin "2")))
    (net (code "3") (name "N$1") (node (ref "D1") (pin "2")) (node (ref "R1") (pin "2")))
    (net (code "4") (name "unconnected-(J1-Pin_3)") (node (ref "J1") (pin "3")))))`

// board writes a .kicad_pcb with the given pads: "REF PAD NET" each.
func board(pads ...string) string {
	byRef := map[string][]string{}
	var refs []string
	for _, p := range pads {
		f := strings.Fields(p)
		if _, ok := byRef[f[0]]; !ok {
			refs = append(refs, f[0])
		}
		net := `(net 0 "")`
		if len(f) > 2 {
			net = `(net 1 "` + f[2] + `")`
		}
		byRef[f[0]] = append(byRef[f[0]], `(pad "`+f[1]+`" smd rect (at 0 0) `+net+`)`)
	}
	var b strings.Builder
	b.WriteString(`(kicad_pcb (version 20241229) (net 0 "")`)
	for _, r := range refs {
		b.WriteString(` (footprint "Lib:FP" (property "Reference" "` + r + `" (at 0 0)) ` + strings.Join(byRef[r], " ") + `)`)
	}
	b.WriteString(`)`)
	return b.String()
}

var good = []string{
	"J1 1 +3V3", "J1 2 GND", "J1 3",
	"R1 1 +3V3", "R1 2 Net-(D1-A)",
	"D1 1 GND", "D1 2 Net-(D1-A)",
}

func compare(t *testing.T, pads []string) string {
	t.Helper()
	n, err := ReadNetlist(netlist)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ReadBoard(board(pads...))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(Compare(n, b), "\n")
}

func TestAgree(t *testing.T) {
	if p := compare(t, good); p != "" {
		t.Errorf("differences, though the nets join the same pins (N$1 is Net-(D1-A)):\n%s", p)
	}
}

func TestDifferences(t *testing.T) {
	replace := func(old, new string) []string {
		var out []string
		for _, p := range good {
			if p == old {
				if new != "" {
					out = append(out, new)
				}
				continue
			}
			out = append(out, p)
		}
		return out
	}
	for _, c := range []struct {
		name string
		pads []string
		want string
	}{
		{"missing net", replace("D1 2 Net-(D1-A)", "D1 2"), "net N$1: pin D1.2 is on no net of the board"},
		{"missing net whole", replace("R1 2 Net-(D1-A)", "R1 2"), "net N$1: pin R1.2 is on no net of the board"},
		{"split", replace("R1 1 +3V3", "R1 1 VCC2"), "net +3V3: split on the board into +3V3 (J1.1), VCC2 (R1.1)"},
		{"joined", replace("D1 1 GND", "D1 1 +3V3"), "board net +3V3: joins netlist nets +3V3, GND"},
		{"extra connection", append(replace("J1 3", "J1 3 GND")), "board net GND: pin J1.3 is on no net of the netlist"},
		{"missing part", replace("D1 1 GND", ""), "net GND: pin D1.1 is on no net of the board"},
	} {
		if p := compare(t, c.pads); !strings.Contains(p, c.want) {
			t.Errorf("%s: want %q in:\n%s", c.name, c.want, p)
		}
	}
	b := []string{"J1 1 +3V3", "J1 2 GND", "R1 1 +3V3", "R1 2 Net-(D1-A)"}
	if p := compare(t, b); !strings.Contains(p, "component D1: in the netlist, not on the board") {
		t.Errorf("no footprint for D1:\n%s", p)
	}
}

func TestReadErrors(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"(kicad_pcb)", "not a KiCad netlist"},
		{"(export (components))", "no (nets ...)"},
		{`(export (nets (net (name "A") (node (ref "R1")))))`, "a node without ref or pin"},
		{"(export (nets", "unclosed list"},
		{`(export "x`, "unclosed string"},
	} {
		if _, err := ReadNetlist(c.src); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("ReadNetlist(%q) = %v, want %q", c.src, err, c.want)
		}
	}
	if _, err := ReadBoard("(export)"); err == nil || !strings.Contains(err.Error(), "not a .kicad_pcb") {
		t.Errorf("ReadBoard: %v", err)
	}
	if _, err := ReadBoard(`(kicad_pcb (footprint "L:F" (pad "1")))`); err == nil || !strings.Contains(err.Error(), "has no reference") {
		t.Errorf("ReadBoard: %v", err)
	}
}

func TestStringEscapes(t *testing.T) {
	n, err := Parse(`(a "x \"y\" z" b "1\n2\t3\\4" "é ü")`)
	if err != nil || n.Arg(0) != `x "y" z` || n.Arg(1) != "b" || n.Arg(2) != "1\n2\t3\\4" || n.Arg(3) != "é ü" {
		t.Errorf("Parse = %q %q %q %q, %v", n.Arg(0), n.Arg(1), n.Arg(2), n.Arg(3), err)
	}
}

func TestParseErrors(t *testing.T) {
	for src, want := range map[string]string{
		"":       "unexpected end",
		"(a) b":  "text after the expression",
		")":      "unexpected )",
		"(a (b)": "unclosed list",
		`(a "b)`: "unclosed string",
	} {
		if _, err := Parse(src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Parse(%q) = %v, want %q", src, err, want)
		}
	}
}

func TestBoardForms(t *testing.T) {
	// KiCad 7's fp_text reference; a net by name alone; (net 0) and (net)
	// for none.
	b, err := ReadBoard(`(kicad_pcb
  (footprint "L:F" (fp_text reference "R1" (at 0 0)) (pad "1" smd (net "VCC")) (pad "2" smd (net 0)) (pad "3" smd (net)))
  (footprint "L:F" (property "Reference" "R2") (pad "1" smd (net 4 "VCC"))))`)
	if err != nil {
		t.Fatal(err)
	}
	if !b.Refs["R1"] || !b.Refs["R2"] {
		t.Errorf("refs %v", b.Refs)
	}
	if got := strings.Join(b.Nets["VCC"], " "); got != "R1.1 R2.1" || len(b.Nets) != 1 {
		t.Errorf("nets %v", b.Nets)
	}
}

func TestPadConflict(t *testing.T) {
	n, _ := ReadNetlist(netlist)
	b, err := ReadBoard(board(append(good, "R1 2 GND")...))
	if err != nil {
		t.Fatal(err)
	}
	if p := strings.Join(Compare(n, b), "\n"); !strings.Contains(p, "pin R1.2: pads on board nets Net-(D1-A) and GND") {
		t.Errorf("a pin on two nets is not reported:\n%s", p)
	}
}

func TestDuplicateNode(t *testing.T) {
	n, err := ReadNetlist(`(export (nets (net (name "A") (node (ref "R1") (pin "1")) (node (ref "R1") (pin "1")))))`)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Nets["A"]) != 1 {
		t.Errorf("a pin listed twice counts twice: %v", n.Nets)
	}
}

func TestVerdict(t *testing.T) {
	for _, c := range []struct {
		problems []string
		want     string
		ok       bool
	}{
		{nil, "", true},
		{[]string{"x"}, "", false},
		{[]string{"net A: x"}, "net A", true},
		{nil, "net A", false},
	} {
		if err := Verdict(c.problems, c.want); (err == nil) != c.ok {
			t.Errorf("Verdict(%q, %q) = %v", c.problems, c.want, err)
		}
	}
}

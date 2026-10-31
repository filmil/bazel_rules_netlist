// SPDX-License-Identifier: Apache-2.0

// Command parity is netlist_parity_test's checker. It reads, from the
// environment:
//
//   - PARITY_NETLIST: the netlist, as a runfiles path (repository/path);
//   - PARITY_BOARD: the .kicad_pcb, the same way;
//   - PARITY_WANT_ERROR: empty when the two must agree; else text that one
//     of the differences must contain (a known-bad board).
package main

import (
	"fmt"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/filmil/bazel_rules_netlist/netlist/private/parity"
)

func read(env string) ([]byte, error) {
	p, err := runfiles.Rlocation(os.Getenv(env))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", env, err)
	}
	return os.ReadFile(p)
}

func run() error {
	nb, err := read("PARITY_NETLIST")
	if err != nil {
		return err
	}
	bb, err := read("PARITY_BOARD")
	if err != nil {
		return err
	}
	netlist, err := parity.ReadNetlist(string(nb))
	if err != nil {
		return err
	}
	board, err := parity.ReadBoard(string(bb))
	if err != nil {
		return err
	}
	problems := parity.Compare(netlist, board)
	for _, p := range problems {
		fmt.Println(p)
	}
	if err := parity.Verdict(problems, os.Getenv("PARITY_WANT_ERROR")); err != nil {
		return fmt.Errorf("%s and %s: %v", os.Getenv("PARITY_NETLIST"), os.Getenv("PARITY_BOARD"), err)
	}
	fmt.Println("as expected")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "netlist_parity_test:", err)
		os.Exit(1)
	}
}

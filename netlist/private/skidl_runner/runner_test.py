# SPDX-License-Identifier: Apache-2.0
"""Tests of the SKiDL runner."""

import os
import sys
import tempfile
import unittest

import runner


def symbol_dir():
    """The pinned KiCad symbol libraries, from the test's data."""
    for p in os.environ["TEST_SYMBOLS"].split():
        if p.endswith(".kicad_sym"):
            return os.path.dirname(os.path.abspath(p))
    raise AssertionError("no .kicad_sym in TEST_SYMBOLS")


class RunnerTest(unittest.TestCase):
    def setUp(self):
        # main() sets the environment, sys.argv, sys.path and SKiDL's
        # default tool; put them back after each test.
        env, argv, path = dict(os.environ), list(sys.argv), list(sys.path)
        import builtins

        import skidl

        # Each test starts with an empty default circuit.
        builtins.default_circuit = skidl.Circuit()

        def restore():
            os.environ.clear()
            os.environ.update(env)
            sys.argv[:] = argv
            sys.path[:] = path

        self.addCleanup(restore)

    def fails(self, argv, want):
        with self.assertRaises(SystemExit) as e:
            runner.main(argv)
        self.assertIn(want, str(e.exception))

    def test_missing_symbol_dir(self):
        self.fails(["--kicad_version", "9", "--symbol_dir", "/no/such/dir", "x.py"], "is not a directory")

    def test_unknown_kicad_version(self):
        with tempfile.TemporaryDirectory() as d:
            self.fails(["--kicad_version", "4", "--symbol_dir", d, "x.py"], "SKiDL has no tool kicad4")

    def test_bad_flags(self):
        self.fails(["--symbol_dir", "/tmp", "x.py"], "--kicad_version is required")
        self.fails(["--kicad_version", "9", "--symbol_dir", "/tmp"], "no SCRIPT")
        self.fails(["--kicad_version"], "--kicad_version needs a value")
        self.fails(["--kicad", "9", "--symbol_dir", "/tmp", "x.py"], "unknown flag --kicad")
        self.fails(["--kicad_version", "9", "--kicad_version", "8", "--symbol_dir", "/tmp", "x.py"], "--kicad_version given twice")
        self.assertEqual(
            runner.parse(["--kicad_version", "9", "--symbol_dir", "d", "--path", "a", "--path", "b", "s.py"])[0]["--path"],
            ["a", "b"],
        )

    def test_parse_keeps_script_args(self):
        self.assertEqual(
            runner.parse(["--kicad_version", "9", "--symbol_dir", "d", "s.py", "--foo", "--kicad_version", "x"]),
            ({"--kicad_version": "9", "--symbol_dir": "d"}, "s.py", ["--foo", "--kicad_version", "x"]),
        )
        self.assertEqual(
            runner.parse(["--kicad_version", "9", "--symbol_dir", "d", "--netlist", "o.net", "--", "-s.py", "a"]),
            ({"--kicad_version": "9", "--symbol_dir": "d", "--netlist": "o.net"}, "-s.py", ["a"]),
        )

    def test_fix_header(self):
        raw = (
            '(export (version "E")\n  (design\n    (source "/tmp/x/c.py")\n    (date "10/31/2026 01:02 AM")\n'
            '    (tool "SKiDL (3.0.0)")\n    (sheet (title_block (date) (source "../../_runner_stage2_bootstrap.py"))))\n'
            '  (components (comp (fields (field (name "SKiDL Line") "../../../../pkg/parts.py:12")))))'
        )
        got = runner.fix_header(raw, "pkg/c.py", "SKiDL (3.0.0)", "/x/execroot/_main")
        self.assertIn('(source "pkg/c.py")', got)
        self.assertIn('(date "1970-01-01T00:00:00+0000")', got)
        self.assertNotIn("/tmp/x", got)
        self.assertNotIn("bootstrap", got)
        self.assertIn("(date)", got)
        self.assertIn('(name "SKiDL Line") "pkg/parts.py:12"', got)
        with self.assertRaises(SystemExit) as e:
            runner.fix_header('(export (design (source "a")) (components))', "a", "t", "/")
        self.assertIn("no (date ...)", str(e.exception))

    def test_runs_script_with_env(self):
        with tempfile.TemporaryDirectory() as d:
            out = os.path.join(d, "out.txt")
            script = os.path.join(d, "s.py")
            with open(os.path.join(d, "helper.py"), "w") as f:
                f.write("NAME = 'helper'\n")
            with open(script, "w") as f:
                f.write(
                    "import os, sys, skidl, helper\n"
                    "open(sys.argv[1], 'w').write(' '.join([os.environ['KICAD9_SYMBOL_DIR'], "
                    "os.environ['KICAD_SYMBOL_DIR'], skidl.get_default_tool(), helper.NAME] + sys.argv[2:]))\n"
                )
            runner.main(["--kicad_version", "9", "--symbol_dir", d, script, out, "--foo", "bar"])
            with open(out) as f:
                d = os.path.abspath(d)
                self.assertEqual(f.read(), " ".join([d, d, "kicad9", "helper", "--foo", "bar"]))


    def test_netlist_needs_parts(self):
        with tempfile.TemporaryDirectory() as d:
            script = os.path.join(d, "empty.py")
            with open(script, "w") as f:
                f.write("import skidl\n")
            self.fails(["--kicad_version", "9", "--symbol_dir", d, "--netlist", os.path.join(d, "o.net"), script], "made no parts")


    def test_workspace_path(self):
        cwd = "/sandbox/7/execroot/_main"
        for given, want in [
            ("../../../../pkg/c.py", "pkg/c.py"),
            ("/sandbox/7/execroot/_main/pkg/c.py", "pkg/c.py"),
            ("bazel-out/k8-fastbuild/bin/pkg/gen.py", "pkg/gen.py"),
            ("../../bazel-out/k8-opt/bin/pkg/gen.py", "pkg/gen.py"),
            ("pkg/c.py", "pkg/c.py"),
        ]:
            self.assertEqual(runner.workspace_path(given, cwd), want, given)

    def test_header_only_in_design(self):
        raw = (
            '(export (design (source "/abs/c.py") (date "now") (tool "x"))\n'
            '  (components (comp (fields (field (name "Note") "x") (source "kept")))))'
        )
        got = runner.fix_header(raw, "c.py", "t", "/")
        self.assertIn('(source "kept")', got)
        self.assertIn('(source "c.py")', got)

    def test_backslash_in_source(self):
        got = runner.fix_header('(export (design (source "a") (date "d") (tool "t")) (components))', "a\\1\\n", "t", "/")
        self.assertIn('(source "a\\1\\n")', got)

    def test_writes_netlist_with_path(self):
        with tempfile.TemporaryDirectory() as d:
            lib = os.path.join(d, "lib")
            os.mkdir(lib)
            with open(os.path.join(lib, "mod.py"), "w") as f:
                f.write("VALUE = '4k7'\n")
            script = os.path.join(d, "c.py")
            with open(script, "w") as f:
                f.write("import mod, skidl\nskidl.Part('Device', 'R', ref='R7', value=mod.VALUE)\n")
            out = os.path.join(d, "o.net")
            runner.main(["--kicad_version", "9", "--symbol_dir", symbol_dir(), "--netlist", out, "--source", "pkg/c.py", "--path", lib, script])
            with open(out) as f:
                net = f.read()
            self.assertIn('(source "pkg/c.py")', net)
            self.assertIn('(ref "R7")', net)
            self.assertIn('(value "4k7")', net)
            self.assertFalse(os.path.exists(out + ".raw"))

    def test_fresh_circuit_each_run(self):
        # A part left in the default circuit by an earlier run in this
        # process does not count for the next one.
        with tempfile.TemporaryDirectory() as d:
            first = os.path.join(d, "first.py")
            with open(first, "w") as f:
                f.write("import skidl\nskidl.Part('Device', 'R', ref='R1')\n")
            runner.main(["--kicad_version", "9", "--symbol_dir", symbol_dir(), first])
            empty = os.path.join(d, "empty.py")
            with open(empty, "w") as f:
                f.write("import skidl\n")
            self.fails(["--kicad_version", "9", "--symbol_dir", symbol_dir(), "--netlist", os.path.join(d, "o.net"), empty], "made no parts")


if __name__ == "__main__":
    unittest.main()

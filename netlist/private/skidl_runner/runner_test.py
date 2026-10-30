# SPDX-License-Identifier: Apache-2.0
"""Tests of the SKiDL runner."""

import os
import sys
import tempfile
import unittest

import runner


class RunnerTest(unittest.TestCase):
    def setUp(self):
        # main() sets the environment, sys.argv, sys.path and SKiDL's
        # default tool; put them back after each test.
        env, argv, path = dict(os.environ), list(sys.argv), list(sys.path)

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

    def test_parse_keeps_script_args(self):
        self.assertEqual(
            runner.parse(["--kicad_version", "9", "--symbol_dir", "d", "s.py", "--foo", "--kicad_version", "x"]),
            ("9", "d", "s.py", ["--foo", "--kicad_version", "x"]),
        )
        self.assertEqual(runner.parse(["--kicad_version", "9", "--symbol_dir", "d", "--", "-s.py", "a"]), ("9", "d", "-s.py", ["a"]))

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


if __name__ == "__main__":
    unittest.main()

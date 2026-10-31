# SPDX-License-Identifier: Apache-2.0
"""Runs a SKiDL circuit script with the pinned KiCad symbol libraries.

    runner --kicad_version N --symbol_dir DIR [--] SCRIPT [ARGS...]

KICAD<N>_SYMBOL_DIR and KICAD_SYMBOL_DIR point at DIR before SKiDL is
imported, and SKiDL's default tool is set to kicad<N>, so
`Part("Device", "R")` finds the symbol in the pinned libraries and
nowhere else. SCRIPT runs as `__main__`, as `python3 SCRIPT ARGS...`
would run it: ARGS are its `sys.argv[1:]`, passed as they are, and its
directory is first on `sys.path`. The working directory is the action's,
inside the sandbox.
"""

import os
import runpy
import sys

USAGE = "usage: runner --kicad_version N --symbol_dir DIR [--] SCRIPT [ARGS...]"


def parse(argv):
    """Returns (kicad_version, symbol_dir, script, args). The runner's own
    flags come first; everything from SCRIPT on is the script's."""
    flags = {}
    i = 0
    while i < len(argv):
        a = argv[i]
        if a == "--":
            i += 1
            break
        if a in ("--kicad_version", "--symbol_dir"):
            if i + 1 >= len(argv):
                sys.exit("runner: %s needs a value\n%s" % (a, USAGE))
            flags[a] = argv[i + 1]
            i += 2
            continue
        if a.startswith("-"):
            sys.exit("runner: unknown flag %s\n%s" % (a, USAGE))
        break
    if i >= len(argv):
        sys.exit("runner: no SCRIPT\n%s" % USAGE)
    for f in ("--kicad_version", "--symbol_dir"):
        if f not in flags:
            sys.exit("runner: %s is required\n%s" % (f, USAGE))
    return flags["--kicad_version"], flags["--symbol_dir"], argv[i], argv[i + 1:]


def main(argv):
    version, symbol_dir, script, args = parse(argv)
    symbol_dir = os.path.abspath(symbol_dir)
    if not os.path.isdir(symbol_dir):
        sys.exit("runner: --symbol_dir %s is not a directory" % symbol_dir)
    for name in ("KICAD%s_SYMBOL_DIR" % version, "KICAD_SYMBOL_DIR"):
        os.environ[name] = symbol_dir

    import skidl

    tool = "kicad%s" % version
    if tool not in skidl.tools.ALL_TOOLS:
        sys.exit("runner: SKiDL has no tool %s; it has %s" % (tool, ", ".join(sorted(skidl.tools.ALL_TOOLS))))
    skidl.set_default_tool(tool)
    sys.argv = [script] + args
    sys.path.insert(0, os.path.dirname(os.path.abspath(script)))
    runpy.run_path(script, run_name="__main__")


if __name__ == "__main__":
    main(sys.argv[1:])

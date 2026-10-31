# SPDX-License-Identifier: Apache-2.0
"""Runs a SKiDL circuit script with the pinned KiCad symbol libraries.

    runner --kicad_version N --symbol_dir DIR [--netlist OUT [--source NAME]]
           [--path DIR]... [--] SCRIPT [ARGS...]

KICAD<N>_SYMBOL_DIR and KICAD_SYMBOL_DIR point at DIR before SKiDL is
imported, and SKiDL's default tool is set to kicad<N>, so
`Part("Device", "R")` finds the symbol in the pinned libraries and
nowhere else. SCRIPT runs as `__main__`, as `python3 SCRIPT ARGS...`
would run it: ARGS are its `sys.argv[1:]`, passed as they are, and its
directory is first on `sys.path`, then each --path, for the modules it
imports (skidl_netlist's deps). The working directory is the action's,
inside the sandbox.

Python's random generator is seeded with 0 before SCRIPT runs: SKiDL tags
a part that has no tag= with random letters, and the netlist's component
time stamps come from the tags, so they are the same on every run.

With --netlist OUT, the runner then writes the default circuit's KiCad
netlist to OUT (skidl_netlist, issue #10), with a fixed header: the source
is NAME (SCRIPT's path in the workspace; SCRIPT itself by default), the date is 1970-01-01, and the tool is SKiDL
with its version. Then the same circuit gives the same bytes.
"""

import os
import random
import runpy
import sys

USAGE = (
    "usage: runner --kicad_version N --symbol_dir DIR [--netlist OUT [--source NAME]] "
    + "[--path DIR]... [--] SCRIPT [ARGS...]"
)

FLAGS = ("--kicad_version", "--symbol_dir", "--netlist", "--source", "--path")
REPEATED = ("--path",)
REQUIRED = ("--kicad_version", "--symbol_dir")

# KiCad's own netlist export writes this date when SOURCE_DATE_EPOCH is 0
# (rules_kicad's kicad_netlist), so both kinds of netlist read the same.
FIXED_DATE = "1970-01-01T00:00:00+0000"


def parse(argv):
    """Returns (flags, script, args): flags maps each of FLAGS given to its
    value. The runner's own flags come first; everything from SCRIPT on is
    the script's."""
    flags = {}
    i = 0
    while i < len(argv):
        a = argv[i]
        if a == "--":
            i += 1
            break
        if a in FLAGS:
            if i + 1 >= len(argv):
                sys.exit("runner: %s needs a value\n%s" % (a, USAGE))
            if a in REPEATED:
                flags.setdefault(a, []).append(argv[i + 1])
            elif a in flags:
                sys.exit("runner: %s given twice\n%s" % (a, USAGE))
            else:
                flags[a] = argv[i + 1]
            i += 2
            continue
        if a.startswith("-"):
            sys.exit("runner: unknown flag %s\n%s" % (a, USAGE))
        break
    if i >= len(argv):
        sys.exit("runner: no SCRIPT\n%s" % USAGE)
    for f in REQUIRED:
        if f not in flags:
            sys.exit("runner: %s is required\n%s" % (f, USAGE))
    return flags, argv[i], argv[i + 1:]


def workspace_path(path, cwd):
    """The path of a file in the workspace, from a path SKiDL wrote: one
    relative to where its caller ran (../ up to the root of the sandbox,
    then the file's path), or an absolute one under the working directory
    (the execution root, without a sandbox). A generated file loses its
    bazel-out/<config>/bin/ too, so the configuration does not show."""
    import re

    if os.path.isabs(path) and path.startswith(cwd.rstrip("/") + "/"):
        path = path[len(cwd.rstrip("/")) + 1:]
    path = re.sub(r"^(?:\.\./)+", "", path)
    return re.sub(r"^bazel-out/[^/]+/bin/", "", path)


def fix_header(text, source, tool, cwd):
    """Makes a netlist the same wherever it is written. In the (design ...)
    header, every (source ...) is `source` (SKiDL also names its own entry
    point in each sheet's title block, a path that depends on where the
    action runs), and the first (date ...) and (tool ...) are fixed. Each
    part's "SKiDL Line" names its file by its workspace path. Fails if a
    header field is missing: SKiDL changed its format, and the netlist
    would not be the same twice."""
    import re

    end = text.find("(components")
    if end < 0:
        sys.exit("runner: the netlist has no (components ...)")
    head, rest = text[:end], text[end:]
    for field, value, count in (("source", source, 0), ("date", FIXED_DATE, 1), ("tool", tool, 1)):
        repl = '(%s "%s")' % (field, value)
        head, n = re.subn(r'\(%s "(?:[^"\\]|\\.)*"\)' % field, lambda _: repl, head, count=count)
        if n == 0:
            sys.exit("runner: the netlist has no (%s ...) in its header" % field)

    def line(m):
        path, _, num = m.group(2).rpartition(":")
        return m.group(1) + workspace_path(path, cwd) + ":" + num + '"'

    rest = re.sub(r'(\(name "SKiDL Line"\) ")([^"]*)"', line, rest)
    return head + rest


def main(argv):
    flags, script, args = parse(argv)
    version, symbol_dir = flags["--kicad_version"], flags["--symbol_dir"]
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
    # SKiDL reads KICAD<N>_SYMBOL_DIR when it is first imported, which may
    # have been before this; set its search path itself, to DIR alone.
    skidl.lib_search_paths[tool] = [symbol_dir]
    import builtins

    # SKiDL adds what Part() and Net() make to builtins.default_circuit;
    # start from an empty one, whatever ran before in this process.
    builtins.default_circuit = skidl.Circuit()
    random.seed(0)
    sys.argv = [script] + args
    sys.path[0:0] = [os.path.dirname(os.path.abspath(script))] + [os.path.abspath(p) for p in flags.get("--path", [])]
    runpy.run_path(script, run_name="__main__")
    if "--netlist" in flags:
        write_netlist(skidl, flags["--netlist"], script, flags.get("--source", script))


def write_netlist(skidl, out, script, source):
    import importlib.metadata

    import builtins

    # SKiDL keeps the circuit that Part() and Net() add to in builtins.
    circuit = builtins.default_circuit
    if not circuit.parts:
        sys.exit("runner: %s made no parts" % script)
    tmp = out + ".raw"
    # do_backup would write a library of the circuit's parts beside it.
    circuit.generate_netlist(file_=tmp, do_backup=False)
    with open(tmp) as f:
        text = f.read()
    os.remove(tmp)
    tool = "SKiDL (%s)" % importlib.metadata.version("skidl")
    with open(out, "w") as f:
        f.write(fix_header(text, source, tool, os.getcwd()))


if __name__ == "__main__":
    main(sys.argv[1:])

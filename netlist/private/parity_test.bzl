# SPDX-License-Identifier: Apache-2.0
"""netlist_parity_test: a board's nets against a netlist's (issue #11)."""

load(":run_tool.bzl", "HELPERS_TOOLCHAIN_TYPE")

def _rlocation(ctx, f):
    """Returns the file's runfiles path, `repository/path`.

    The checker resolves it with the runfiles library, so files of other
    repositories work too.
    """
    if f.short_path.startswith("../"):
        return f.short_path[3:]
    return ctx.workspace_name + "/" + f.short_path

def _netlist_parity_test_impl(ctx):
    parity = ctx.toolchains[HELPERS_TOOLCHAIN_TYPE].helpers.parity
    exe = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(
        output = exe,
        target_file = parity.files_to_run.executable,
        is_executable = True,
    )
    return [
        DefaultInfo(
            executable = exe,
            runfiles = ctx.runfiles(files = [ctx.file.netlist, ctx.file.board]).merge(parity.default_runfiles),
        ),
        RunEnvironmentInfo(environment = {
            "PARITY_BOARD": _rlocation(ctx, ctx.file.board),
            "PARITY_NETLIST": _rlocation(ctx, ctx.file.netlist),
            "PARITY_WANT_ERROR": ctx.attr.want_error,
        }),
    ]

netlist_parity_test = rule(
    implementation = _netlist_parity_test_impl,
    test = True,
    doc = """Checks that a board's nets join the pins a netlist's nets join.

```starlark
netlist_parity_test(
    name = "parity_test",
    size = "small",
    board = "main.kicad_pcb",
    netlist = ":netlist",   # skidl_netlist, or hdlf_rules_kicad's kicad_netlist
)
```

Nets are compared by their pins, not their names: SKiDL names an unnamed
net `N$1` where KiCad's schematic says `Net-(D1-A)`, and the two match when
they join the same pins. The test fails, naming each difference, when:

* a netlist net's pin is on no net of the board (a missing net or pad);
* a netlist net is split over two or more nets of the board;
* a board net joins pins of two or more netlist nets, or a pin the
  netlist leaves unconnected;
* a component of the netlist has no footprint on the board.

A pin whose pads (two of one number, such as an SOT-223's tab) are on two
nets fails too. A netlist net of one pin connects nothing and is not
checked, and neither is a board net of one pad: a footprint the netlist
does not name, such as a mounting hole, fails only when its pads are on a
net with other pins.
""",
    attrs = {
        "netlist": attr.label(
            doc = "The KiCad netlist (`skidl_netlist` or `kicad_netlist`).",
            mandatory = True,
            allow_single_file = [".net"],
        ),
        "board": attr.label(
            doc = "The board, a `.kicad_pcb` file.",
            mandatory = True,
            allow_single_file = [".kicad_pcb"],
        ),
        "want_error": attr.string(
            doc = "For a known-bad board: text that one of the differences must contain. The " +
                  "test then passes only if the check finds it. Empty by default: the two must " +
                  "agree.",
        ),
    },
    toolchains = [HELPERS_TOOLCHAIN_TYPE],
)

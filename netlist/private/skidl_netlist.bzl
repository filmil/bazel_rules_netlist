# SPDX-License-Identifier: Apache-2.0
"""skidl_netlist: a KiCad netlist from a SKiDL circuit script (issue #10)."""

load(":run_tool.bzl", "RUN_TOOL_TOOLCHAINS", "run_tool")

SKIDL_TOOLCHAIN_TYPE = Label("//netlist:skidl_toolchain_type")

def _skidl_netlist_impl(ctx):
    skidl = ctx.toolchains[SKIDL_TOOLCHAIN_TYPE].skidl
    out = ctx.actions.declare_file(ctx.label.name + ".net")
    deps = ctx.files.deps
    args = ctx.actions.args()
    args.add("--kicad_version", skidl.kicad_version)
    args.add("--symbol_dir", skidl.symbol_dir)
    args.add("--netlist", out)
    args.add("--source", ctx.file.src.short_path)

    # In the order deps gives them, each directory once.
    dirs = []
    for f in deps:
        if f.dirname not in dirs:
            dirs.append(f.dirname)
    for d in dirs:
        args.add("--path", d)
    args.add("--")
    args.add(ctx.file.src)
    run_tool(
        ctx,
        SKIDL_TOOLCHAIN_TYPE,
        "skidl",
        arguments = [args],
        inputs = depset([ctx.file.src] + deps, transitive = [skidl.symbols]),
        outputs = [out],
        mnemonic = "SkidlNetlist",
        progress_message = "Running the SKiDL circuit %{label}",
    )
    return [DefaultInfo(files = depset([out]))]

skidl_netlist = rule(
    implementation = _skidl_netlist_impl,
    doc = """A KiCad netlist, `<name>.net`, from a SKiDL circuit script.

```starlark
load("@hdlf_rules_netlist//netlist:defs.bzl", "skidl_netlist")
load("@hdlf_rules_kicad//kicad:defs.bzl", "kicad_pcb")

skidl_netlist(
    name = "netlist",
    src = "circuit.py",
    deps = ["parts.py"],   # modules circuit.py imports
)

kicad_pcb(
    name = "pcb",
    board = "main.kicad_pcb",
    netlist = ":netlist",
)
```

The script builds its circuit with SKiDL; it need not call
`generate_netlist()`. Its parts come from the KiCad symbol libraries of
the SKiDL toolchain (KiCad 9.0.9, the libraries hdlf_rules_kicad pins),
and nowhere else. The netlist is the default circuit's, as SKiDL writes it
for KiCad, with a fixed header: the source is the script's path, the date
1970-01-01 and the tool SKiDL with its version. So the same script gives
the same bytes, wherever it is built and whether or not the script is
generated. A script that makes no parts fails.

SKiDL tags each part, and the netlist's time stamp of a component, which
links it to its footprint on the board, comes from the tag. A part with no
`tag=` gets one from a random generator that the runner seeds, so the
same script gives the same tags; but adding or moving a part can change
the tags of the parts made after it. Give `tag=` to a part whose board
link must stay put.
""",
    attrs = {
        "src": attr.label(
            doc = "The circuit script.",
            mandatory = True,
            allow_single_file = [".py"],
        ),
        "deps": attr.label_list(
            doc = "Python modules the script imports, as plain .py files: each file's " +
                  "directory goes on `sys.path`, after the script's own, in this order.",
            allow_files = [".py"],
        ),
    },
    toolchains = RUN_TOOL_TOOLCHAINS + [SKIDL_TOOLCHAIN_TYPE],
)

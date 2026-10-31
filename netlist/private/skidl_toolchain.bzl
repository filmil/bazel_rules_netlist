# SPDX-License-Identifier: Apache-2.0
"""The SKiDL toolchain (issue #9)."""

SkidlToolchainInfo = provider(
    doc = "What the SKiDL toolchain gives: the KiCad symbol libraries SKiDL reads.",
    fields = {
        "kicad_version": "The KiCad major version of the libraries, such as \"9\".",
        "symbols": "depset of the libraries' .kicad_sym files, all in one directory.",
        "symbol_dir": "The path of that directory, for the runner's --symbol_dir.",
    },
)

def _skidl_toolchain_impl(ctx):
    symbols = [f for f in ctx.files.symbols if f.extension == "kicad_sym"]
    if not symbols:
        fail("%s: symbols gives no .kicad_sym files" % ctx.label)
    dirs = {f.dirname: True for f in symbols}
    if len(dirs) != 1:
        fail("%s: the .kicad_sym files of symbols must be in one directory, not %s" % (ctx.label, sorted(dirs.keys())))
    return [platform_common.ToolchainInfo(
        tools = {"skidl": ctx.attr.runner[DefaultInfo].files_to_run},
        skidl = SkidlToolchainInfo(
            kicad_version = ctx.attr.kicad_version,
            symbols = depset(symbols),
            symbol_dir = dirs.keys()[0],
        ),
    )]

skidl_toolchain = rule(
    implementation = _skidl_toolchain_impl,
    doc = """A SKiDL toolchain: the runner and the KiCad symbol libraries.

```starlark
skidl_toolchain(
    name = "skidl_impl",
    runner = "//netlist/private/skidl_runner:runner",
    symbols = "@kicad_libs//:symbols",
)

toolchain(
    name = "skidl",
    toolchain = ":skidl_impl",
    toolchain_type = "@hdlf_rules_netlist//netlist:skidl_toolchain_type",
)
```

Rules run the runner by the tool name `skidl` through `run_tool`:

```
runner --kicad_version N --symbol_dir DIR SCRIPT [ARGS...]
```

with N and DIR from `SkidlToolchainInfo` (`kicad_version`, `symbol_dir`).
It points KICAD<N>_SYMBOL_DIR and KICAD_SYMBOL_DIR at DIR, sets SKiDL's
default tool to kicad<N>, and runs SCRIPT as `__main__` with ARGS as they
are (`--` before SCRIPT if it starts with `-`).
""",
    attrs = {
        "runner": attr.label(
            doc = "The SKiDL runner: `runner --kicad_version N --symbol_dir DIR [--] SCRIPT [ARGS...]`.",
            mandatory = True,
            executable = True,
            cfg = "exec",
        ),
        "symbols": attr.label(
            doc = "The KiCad symbol libraries: the .kicad_sym files among these, all in one " +
                  "directory. Other files are left out.",
            mandatory = True,
            allow_files = True,
        ),
        "kicad_version": attr.string(
            doc = "The KiCad major version of the libraries.",
            default = "9",
        ),
    },
)

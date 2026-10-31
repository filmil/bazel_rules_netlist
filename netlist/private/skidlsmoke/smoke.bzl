# SPDX-License-Identifier: Apache-2.0
"""Runs a circuit script with the SKiDL toolchain, for its smoke test."""

load("//netlist/private:run_tool.bzl", "RUN_TOOL_TOOLCHAINS", "run_tool")

SKIDL_TOOLCHAIN_TYPE = Label("//netlist:skidl_toolchain_type")

def _skidl_smoke_impl(ctx):
    skidl = ctx.toolchains[SKIDL_TOOLCHAIN_TYPE].skidl
    out = ctx.actions.declare_file(ctx.label.name + ".txt")
    run_tool(
        ctx,
        SKIDL_TOOLCHAIN_TYPE,
        "skidl",
        arguments = [
            "--kicad_version",
            skidl.kicad_version,
            "--symbol_dir",
            skidl.symbol_dir,
            ctx.file.script.path,
            out.path,
        ],
        inputs = depset([ctx.file.script], transitive = [skidl.symbols]),
        outputs = [out],
        mnemonic = "SkidlSmoke",
    )
    return [DefaultInfo(files = depset([out]))]

skidl_smoke = rule(
    implementation = _skidl_smoke_impl,
    attrs = {"script": attr.label(mandatory = True, allow_single_file = [".py"])},
    toolchains = RUN_TOOL_TOOLCHAINS + [SKIDL_TOOLCHAIN_TYPE],
)

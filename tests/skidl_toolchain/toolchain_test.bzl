# SPDX-License-Identifier: Apache-2.0
"""Analysis tests of skidl_toolchain (issue #9)."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

def _fails_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, ctx.attr.message)
    return analysistest.end(env)

# The target fails analysis with a message that contains `message`.
fails_test = analysistest.make(
    _fails_impl,
    expect_failure = True,
    attrs = {"message": attr.string(mandatory = True)},
)

def _info_impl(ctx):
    env = analysistest.begin(ctx)
    info = analysistest.target_under_test(env)[platform_common.ToolchainInfo]
    skidl = info.skidl
    asserts.equals(env, ctx.attr.kicad_version, skidl.kicad_version)
    asserts.equals(env, ctx.attr.symbol_dir, skidl.symbol_dir)
    asserts.equals(env, sorted(ctx.attr.symbols), sorted([f.basename for f in skidl.symbols.to_list()]))
    asserts.true(env, "skidl" in info.tools, "no tool skidl")
    return analysistest.end(env)

# The toolchain's SkidlToolchainInfo and tools.
info_test = analysistest.make(
    _info_impl,
    attrs = {
        "kicad_version": attr.string(mandatory = True),
        "symbol_dir": attr.string(mandatory = True),
        "symbols": attr.string_list(mandatory = True),
    },
)

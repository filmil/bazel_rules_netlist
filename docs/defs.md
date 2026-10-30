<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of hdlf_rules_netlist.

Load every rule from this file:

```starlark
load("@hdlf_rules_netlist//netlist:defs.bzl", "determinism_test", "swap_test", "tool_toolchain")
```

The rules arrive in milestone M6 (see the issues of
https://github.com/filmil/bazel_rules_netlist). This file is the only public
entry point; everything under `//netlist/private` may change without notice.

<a id="skidl_toolchain"></a>

## skidl_toolchain

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "skidl_toolchain")

skidl_toolchain(<a href="#skidl_toolchain-name">name</a>, <a href="#skidl_toolchain-kicad_version">kicad_version</a>, <a href="#skidl_toolchain-runner">runner</a>, <a href="#skidl_toolchain-symbols">symbols</a>)
</pre>

A SKiDL toolchain: the runner and the KiCad symbol libraries.

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

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="skidl_toolchain-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="skidl_toolchain-kicad_version"></a>kicad_version |  The KiCad major version of the libraries.   | String | optional |  `"9"`  |
| <a id="skidl_toolchain-runner"></a>runner |  The SKiDL runner: `runner --kicad_version N --symbol_dir DIR [--] SCRIPT [ARGS...]`.   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |
| <a id="skidl_toolchain-symbols"></a>symbols |  The KiCad symbol libraries: the .kicad_sym files among these, all in one directory. Other files are left out.   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |


<a id="tool_toolchain"></a>

## tool_toolchain

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "tool_toolchain")

tool_toolchain(<a href="#tool_toolchain-name">name</a>, <a href="#tool_toolchain-tools">tools</a>)
</pre>

A toolchain that offers tools by name, for `run_tool`.

Example, a toolchain for another build of the SKiDL runner (the toolchain
type arrives with this module's toolchain issue):

```starlark
tool_toolchain(
    name = "my_skidl_impl",
    tools = {"//third_party/skidl:skidl": "skidl"},
)

toolchain(
    name = "my_skidl",
    toolchain = ":my_skidl_impl",
    toolchain_type = "@hdlf_rules_netlist//netlist:skidl_toolchain_type",
)
```

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="tool_toolchain-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="tool_toolchain-tools"></a>tools |  Executable targets, each mapped to the name rules use for it.   | <a href="https://bazel.build/rules/lib/core/dict">Dictionary: Label -> String</a> | required |  |


<a id="SkidlToolchainInfo"></a>

## SkidlToolchainInfo

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "SkidlToolchainInfo")

SkidlToolchainInfo(<a href="#SkidlToolchainInfo-kicad_version">kicad_version</a>, <a href="#SkidlToolchainInfo-symbols">symbols</a>, <a href="#SkidlToolchainInfo-symbol_dir">symbol_dir</a>)
</pre>

What the SKiDL toolchain gives: the KiCad symbol libraries SKiDL reads.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="SkidlToolchainInfo-kicad_version"></a>kicad_version |  The KiCad major version of the libraries, such as "9".    |
| <a id="SkidlToolchainInfo-symbols"></a>symbols |  depset of the libraries' .kicad_sym files, all in one directory.    |
| <a id="SkidlToolchainInfo-symbol_dir"></a>symbol_dir |  The path of that directory, for the runner's --symbol_dir.    |


<a id="determinism_test"></a>

## determinism_test

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "determinism_test")

determinism_test(<a href="#determinism_test-name">name</a>, <a href="#determinism_test-target">target</a>, <a href="#determinism_test-kwargs">**kwargs</a>)
</pre>

Checks that two builds of `target` give byte-identical outputs.

The target is built twice in one build, under two configurations that
differ only in a setting no rule reads, so every action runs twice.
The test compares the SHA-256 of every output file, directories walked,
and fails on any difference (spec D1).

Two runs in the same second can hide a time stamp with a one-second
resolution, so pair this with a header scan where a format has one. CI
also builds every determinism_test target in two fresh output bases and
compares the results.

Example:

```starlark
determinism_test(
    name = "netlist_determinism_test",
    target = ":netlist",
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="determinism_test-name"></a>name |  the test's name.   |  none |
| <a id="determinism_test-target"></a>target |  the target whose outputs must not change between builds.   |  none |
| <a id="determinism_test-kwargs"></a>kwargs |  common test attributes, such as `size` and `tags`.   |  none |


<a id="swap_test"></a>

## swap_test

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "swap_test")

swap_test(<a href="#swap_test-name">name</a>, <a href="#swap_test-targets">targets</a>, <a href="#swap_test-kwargs">**kwargs</a>)
</pre>

Builds targets with every tool swapped for a stub, and checks them.

The targets are built with `--@hdlf_rules_netlist//netlist/toolchains:use_stub`,
so each tool toolchain that honors the flag resolves to the stub tool.
The stub writes a record into every output, files and directories
alike. The test fails if any output file of the targets lacks the
record: some action made it without a toolchain tool.

Example:

```starlark
swap_test(
    name = "swap_test",
    targets = [":gerbers", ":bom"],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="swap_test-name"></a>name |  the test's name.   |  none |
| <a id="swap_test-targets"></a>targets |  targets whose outputs must all come from the stub.   |  none |
| <a id="swap_test-kwargs"></a>kwargs |  common test attributes, such as `size` and `tags`.   |  none |



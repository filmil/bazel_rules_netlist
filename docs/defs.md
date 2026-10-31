<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of hdlf_rules_netlist.

Load every rule from this file:

```starlark
load(
    "@hdlf_rules_netlist//netlist:defs.bzl",
    "SkidlToolchainInfo",
    "determinism_test",
    "netlist_parity_test",
    "skidl_netlist",
    "skidl_toolchain",
    "swap_test",
    "tool_toolchain",
)
```

The rules arrive in milestone M6 (see the issues of
https://github.com/filmil/bazel_rules_netlist). This file is the only public
entry point; everything under `//netlist/private` may change without notice.

<a id="netlist_parity_test"></a>

## netlist_parity_test

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "netlist_parity_test")

netlist_parity_test(<a href="#netlist_parity_test-name">name</a>, <a href="#netlist_parity_test-board">board</a>, <a href="#netlist_parity_test-netlist">netlist</a>, <a href="#netlist_parity_test-want_error">want_error</a>)
</pre>

Checks that a board's nets join the pins a netlist's nets join.

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

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="netlist_parity_test-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="netlist_parity_test-board"></a>board |  The board, a `.kicad_pcb` file.   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |
| <a id="netlist_parity_test-netlist"></a>netlist |  The KiCad netlist (`skidl_netlist` or `kicad_netlist`).   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |
| <a id="netlist_parity_test-want_error"></a>want_error |  For a known-bad board: text that one of the differences must contain. The test then passes only if the check finds it. Empty by default: the two must agree.   | String | optional |  `""`  |


<a id="skidl_netlist"></a>

## skidl_netlist

<pre>
load("@hdlf_rules_netlist//netlist:defs.bzl", "skidl_netlist")

skidl_netlist(<a href="#skidl_netlist-name">name</a>, <a href="#skidl_netlist-deps">deps</a>, <a href="#skidl_netlist-src">src</a>)
</pre>

A KiCad netlist, `<name>.net`, from a SKiDL circuit script.

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

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="skidl_netlist-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="skidl_netlist-deps"></a>deps |  Python modules the script imports, as plain .py files: each file's directory goes on `sys.path`, after the script's own, in this order.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="skidl_netlist-src"></a>src |  The circuit script.   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |


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



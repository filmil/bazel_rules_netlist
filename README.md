<!-- SPDX-License-Identifier: Apache-2.0 -->
[![Test](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml)
[![Tag and Release](https://github.com/filmil/bazel_rules_netlist/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/releases)
[![Publish to my Bazel registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml)

# hdlf_rules_netlist

Hermetic Bazel rules for circuits written as code.
The build runs a [SKiDL](https://github.com/devbisme/skidl) circuit
script, writes a KiCad netlist that the
[hdlf_rules_kicad](https://github.com/filmil/bazel_rules_kicad) board
rules use, and checks that a board's nets match it.
Every tool comes from a Bazel toolchain, so nothing needs to be installed
on the machine, and two builds give byte-identical outputs.

Only Linux x86_64 hosts are supported in v1. Elsewhere the build fails at
analysis time with a message that says so.

The rules are documented in [docs/defs.md](docs/defs.md), generated from
their sources.

## Using it

`MODULE.bazel`:

```starlark
bazel_dep(name = "hdlf_rules_netlist", version = "0.1.0")
bazel_dep(name = "hdlf_rules_kicad", version = "0.1.0")
```

That is for a release. hdlf_rules_netlist builds on hdlf_rules_kicad,
which is in no registry yet, and no release of hdlf_rules_netlist is made
until it is (issue #8). To use hdlf_rules_netlist meanwhile, pin a commit
of it with an `archive_override(module_name = "hdlf_rules_netlist", ...)`
of GitHub's archive of that commit, and pin hdlf_rules_kicad as
hdlf_rules_netlist does, in your `MODULE.bazel` (an override counts only
in the root module). Each is an archive checked by its hash, so the build
needs no `git` on the host:

```starlark
bazel_dep(name = "hdlf_rules_kicad", version = "0.0.0")
archive_override(
    module_name = "hdlf_rules_kicad",
    integrity = "sha256-bz9RwwNoS0kCPESYe+qbysvlvuv9rVQpqbeqZ+Qj91U=",
    strip_prefix = "bazel_rules_kicad-ecd6ae9329a7830c8f7144db88b5d2bbb24041e9",
    urls = ["https://github.com/filmil/bazel_rules_kicad/archive/ecd6ae9329a7830c8f7144db88b5d2bbb24041e9.tar.gz"],
)
```

Under Bazel 8, rules_python asks for a C++ toolchain even though nothing
is compiled, so a build that finds none on the host must bring one. The
self-contained zig toolchain is enough; Bazel 9 does not ask:

```starlark
bazel_dep(name = "hermetic_cc_toolchain", version = "4.3.0")

zig_toolchains = use_extension("@hermetic_cc_toolchain//toolchain:ext.bzl", "toolchains")
use_repo(zig_toolchains, "zig_sdk")

register_toolchains("@zig_sdk//toolchain:linux_amd64_gnu.2.31")
```

`.bazelrc`, while the module is only in the filmil registry:

```
common --registry=https://raw.githubusercontent.com/filmil/bazel-registry/main
common --registry=https://bcr.bazel.build
```

SKiDL needs no Python on the host and no flag in your `.bazelrc`: it runs
on rules_python's hermetic CPython 3.12, with the wheels that
`requirements_lock.txt` pins by hash and the KiCad 9.0.9 symbol libraries
that hdlf_rules_kicad pins.

## From a circuit script to a checked board

`integration/skidl_netlist` is this whole path, as a user writes it. A
circuit script builds its circuit with SKiDL. It need not call
`generate_netlist()`; the rule writes the netlist:

```python
# circuit.py: a LED and its resistor on a 3.3 V connector.
from skidl import Net, Part

vcc = Net("+3V3")
gnd = Net("GND")

j1 = Part("Connector_Generic", "Conn_01x02", ref="J1", value="Power",
          footprint="Connector_PinHeader_2.54mm:PinHeader_1x02_P2.54mm_Vertical")
r1 = Part("Device", "R", ref="R1", value="1k",
          footprint="Resistor_SMD:R_0603_1608Metric")
d1 = Part("Device", "LED", ref="D1", value="LED",
          footprint="LED_SMD:LED_0603_1608Metric")

vcc += j1[1], r1[1]
gnd += j1[2], d1["K"]
r1[2] += d1["A"]
```

Parts come from the SKiDL toolchain's symbol libraries, and from nowhere
else: by default, the KiCad 9.0.9 libraries that hdlf_rules_kicad pins
(see "Toolchains" below to bring others).
Then the BUILD file:

```starlark
load("@hdlf_rules_kicad//kicad:defs.bzl", "kicad_gerbers", "kicad_pcb")
load("@hdlf_rules_netlist//netlist:defs.bzl", "determinism_test", "netlist_parity_test", "skidl_netlist")

# circuit.py -> netlist.net
skidl_netlist(
    name = "netlist",
    src = "circuit.py",
    deps = ["parts.py"],  # modules circuit.py imports, if any
)

# The board, laid out in KiCad against that netlist.
kicad_pcb(
    name = "pcb",
    board = "main.kicad_pcb",
    netlist = ":netlist",
)

# The board joins exactly the pins the netlist joins.
netlist_parity_test(
    name = "parity_test",
    size = "small",
    board = "main.kicad_pcb",
    netlist = ":netlist",
)

# The same netlist on two builds.
determinism_test(
    name = "netlist_determinism_test",
    size = "small",
    target = ":netlist",
)

# Fab outputs, from hdlf_rules_kicad.
kicad_gerbers(
    name = "gerbers",
    pcb = ":pcb",
)
```

`bazel test //...` runs the SKiDL script, checks the board against the
netlist, and checks that the netlist is the same twice.

## The rules

Load every rule from `@hdlf_rules_netlist//netlist:defs.bzl`.
[docs/defs.md](docs/defs.md) has every attribute.

### skidl_netlist

A KiCad netlist, `<name>.net`, from a SKiDL circuit script, for
`kicad_pcb(netlist = ...)`. The netlist is the same on every build: the
runner fixes its header (the source is the script's workspace path, the
date 1970-01-01, the tool "SKiDL (3.0.0)") and each part's file reference,
and seeds the random generator that SKiDL tags parts with. A component's
time stamp, which links it to its footprint on the board, comes from its
tag, so give `tag=` to a part whose board link must stay put when you add
or move parts before it:

```python
r1 = Part("Device", "R", ref="R1", value="1k", tag="r1")
```

A script that makes no parts fails.

### netlist_parity_test

Checks that a board's nets join the pins a netlist's nets join. Nets are
compared by their pins, not their names, so SKiDL's `N$1` and a
schematic's `Net-(D1-A)` match when they join the same pins. It fails,
naming each difference, on a missing net or pad, a net split over two
board nets, two netlist nets shorted on the board, a pin with pads on two
nets, or a component with no footprint. It reads netlists of
`skidl_netlist` and of hdlf_rules_kicad's `kicad_netlist`.

`want_error` turns it into a known-bad test, for a board that must fail:

```starlark
netlist_parity_test(
    name = "missing_net_test",
    size = "small",
    board = "missing_net.kicad_pcb",
    netlist = ":netlist",
    want_error = "pin D1.2 is on no net of the board",
)
```

### determinism_test

Builds a target twice, in two configurations that change nothing but the
output paths, and fails if any output file differs:

```starlark
determinism_test(
    name = "netlist_determinism_test",
    size = "small",
    target = ":netlist",
)
```

### swap_test

Builds targets with every tool swapped for a stub, and checks that each
output came from the toolchain's tool (spec "Toolchains", requirement 7):

```starlark
swap_test(
    name = "swap_test",
    size = "small",
    targets = [":netlist"],
)
```

### Toolchains: skidl_toolchain, tool_toolchain, SkidlToolchainInfo

The module registers its SKiDL toolchain; most users need nothing more.
To bring other symbol libraries, or another build of the SKiDL runner,
define your own in a BUILD file, here `//toolchains/BUILD.bazel`:

```starlark
load("@hdlf_rules_netlist//netlist:defs.bzl", "skidl_toolchain")

skidl_toolchain(
    name = "my_skidl_impl",
    runner = "@hdlf_rules_netlist//netlist:skidl_runner",
    symbols = "//third_party/kicad:symbols",  # .kicad_sym files in one directory
)

toolchain(
    name = "my_skidl",
    toolchain = ":my_skidl_impl",
    toolchain_type = "@hdlf_rules_netlist//netlist:skidl_toolchain_type",
)
```

and register it in your `MODULE.bazel`. The root module's toolchains come
before those of its dependencies, so Bazel picks yours:

```starlark
register_toolchains("//toolchains:my_skidl")
```

`SkidlToolchainInfo` is what that toolchain gives a rule: the KiCad
version and the symbol libraries. `tool_toolchain` makes a toolchain of
named tools for `run_tool`, the one place where this module's rules start
a tool.

## Schematics from code

A SKiDL circuit gives a netlist, not a `.kicad_sch`, so it gets no ERC and
no schematic renders. For now that stays so: SKiDL 3.0.0 writes KiCad 9
and 10 schematics, but they differ from build to build (random UUIDs,
today's date, and wire routing that changes even with every seed fixed),
and the sheet it wrote fails hdlf_rules_kicad's ERC. Issue #12 has the
details; it can be looked at again when SKiDL's schematizer is
deterministic. Until then, `netlist_parity_test` against the board, with
hdlf_rules_kicad's DRC and board renders, checks a circuit written as code.

## Building

```sh
bazel build //...
bazel test //...
cd integration && bazel test //...
```

After changing a `.bzl` file, regenerate the documentation with
`bazel run //docs:update_docs`.
Format with `bazel run //tools:buildifier`; CI fails when
`bazel run //tools:buildifier_check` or the hermeticity lint
(`bazel run //tools/lint`) does.
`bazel run //tools:requirements.update` writes `requirements_lock.txt`
from `requirements.in`; `bazel run //tools:mirror` checks the mirrored
wheels (`tools/mirror/README.md`).

## Releasing

The operator runs the `Tag and Release` workflow
(`.github/workflows/tag-and-release.yml`, `workflow_dispatch`).
It picks the next version from the conventional commits since the last
tag, makes the tag, the release and its source archive, and opens one
pull request on the filmil registry and one on the Bazel Central Registry.
It needs the repository secret `BCR_PUBLISH_TOKEN`.

None of the modules published here are published with any sort of guarantee
whatsoever. Use at your own risk.

See the overall description at the [project page][pp].

[pp]: https://hdlfactory.com/post/2025/09/29/getting-ready-for-the-brave-new-bazel-modules-world/

## Licenses

This repository is under the Apache License 2.0 (`LICENSE`). The build
downloads the prebuilts below, each pinned by its hash. They are tools of
the build: none of them is copied into the netlists or checks it writes.

**Mirrored here.** PyPI has only the sdists of these pure-Python packages;
their wheels are built from those, as `tools/mirror/README.md` says, and
mirrored in this repository's release
[mirror-skidl-3.0.0](https://github.com/filmil/bazel_rules_netlist/releases/tag/mirror-skidl-3.0.0).

| Package | Version | License | Source |
|---|---|---|---|
| skidl | 3.0.0 | MIT | https://github.com/devbisme/skidl |
| kinet2pcb | 1.1.4 | MIT | https://github.com/devbisme/kinet2pcb |
| hierplace | 1.1.0 | MIT | https://github.com/devbisme/HierPlace |

**Wheels from PyPI**, pinned by hash in `requirements_lock.txt`:

| Package | Version | License | Source |
|---|---|---|---|
| cffi | 2.1.1 | MIT-0 | https://pypi.org/project/cffi/2.1.1/ |
| contourpy | 1.4.0 | BSD-3-Clause | https://pypi.org/project/contourpy/1.4.0/ |
| cycler | 0.12.1 | BSD | https://pypi.org/project/cycler/0.12.1/ |
| deprecation | 2.1.0 | Apache-2.0 | https://pypi.org/project/deprecation/2.1.0/ |
| diskcache | 5.6.3 | Apache-2.0 | https://pypi.org/project/diskcache/5.6.3/ |
| fonttools | 4.68.0 | MIT | https://pypi.org/project/fonttools/4.68.0/ |
| graphviz | 0.21 | MIT | https://pypi.org/project/graphviz/0.21/ |
| h5py | 3.16.0 | BSD-3-Clause | https://pypi.org/project/h5py/3.16.0/ |
| inspice | 1.7.0.7 | GPL-3.0-or-later | https://pypi.org/project/inspice/1.7.0.7/ |
| kiwisolver | 1.5.1 | BSD | https://pypi.org/project/kiwisolver/1.5.1/ |
| markdown-it-py | 4.2.0 | MIT | https://pypi.org/project/markdown-it-py/4.2.0/ |
| matplotlib | 3.11.2 | PSF-based (matplotlib license) | https://pypi.org/project/matplotlib/3.11.2/ |
| mdurl | 0.1.2 | MIT | https://pypi.org/project/mdurl/0.1.2/ |
| numpy | 2.5.4 | BSD-3-Clause AND 0BSD AND MIT AND Zlib AND CC0-1.0 | https://pypi.org/project/numpy/2.5.4/ |
| packaging | 26.3 | Apache-2.0 OR BSD-2-Clause | https://pypi.org/project/packaging/26.3/ |
| pillow | 13.0.0 | MIT-CMU | https://pypi.org/project/pillow/13.0.0/ |
| ply | 3.11 | BSD | https://pypi.org/project/ply/3.11/ |
| pycparser | 3.12 | BSD-3-Clause | https://pypi.org/project/pycparser/3.12/ |
| pygments | 2.21.0 | BSD-2-Clause | https://pypi.org/project/pygments/2.21.0/ |
| pyparsing | 3.3.3 | MIT | https://pypi.org/project/pyparsing/3.3.3/ |
| python-dateutil | 3.0.0 | Apache-2.0 AND BSD-3-Clause | https://pypi.org/project/python-dateutil/3.0.0/ |
| pyyaml | 6.0.3 | MIT | https://pypi.org/project/pyyaml/6.0.3/ |
| rich | 15.0.0 | MIT | https://pypi.org/project/rich/15.0.0/ |
| schematizer | 0.2.0 | MIT | https://pypi.org/project/schematizer/0.2.0/ |
| simp-sexp | 0.3.1 | MIT | https://pypi.org/project/simp-sexp/0.3.1/ |
| six | 1.17.0 | MIT | https://pypi.org/project/six/1.17.0/ |
| tzdata | 2026.6 | Apache-2.0 | https://pypi.org/project/tzdata/2026.6/ |

inspice is GPL-3.0-or-later. SKiDL 3.0.0 requires it on Python 3.11 and
later; it is fetched from PyPI at build time and runs only inside the
build, and nothing of it is in this repository's sources, releases or
outputs.

**Other prebuilts:**

| What | Version | License | Source |
|---|---|---|---|
| CPython, from rules_python 2.4.6 | 3.12.16 | PSF-2.0 | https://github.com/astral-sh/python-build-standalone |
| KiCad symbol libraries, through hdlf_rules_kicad | 9.0.9 | CC-BY-SA-4.0 with the KiCad libraries exception | https://gitlab.com/kicad/libraries/kicad-symbols |
| Go SDK, from rules_go 0.64.2, for the Go helpers | 1.26.7 | BSD-3-Clause | https://go.dev/dl/ |

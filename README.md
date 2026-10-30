<!-- SPDX-License-Identifier: Apache-2.0 -->
[![Test](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml)
[![Tag and Release](https://github.com/filmil/bazel_rules_netlist/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/releases)
[![Publish to my Bazel registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml)

# hdlf_rules_netlist

Hermetic Bazel rules for circuits written as code.
The build runs a SKiDL circuit script, writes a KiCad netlist that the
hdlf_rules_kicad board rules use, and checks that a board's nets match it.
Every tool comes from a Bazel toolchain, so nothing needs to be installed
on the machine, and two builds give byte-identical outputs.

This repository is new. The rules arrive in milestones; see the
[issues](https://github.com/filmil/bazel_rules_netlist/issues) and the
[milestones](https://github.com/filmil/bazel_rules_netlist/milestones).
Only Linux x86_64 hosts are supported.

## Using it

```starlark
bazel_dep(name = "hdlf_rules_netlist", version = "0.1.0")
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

`.bazelrc`, while the module is only in the filmil registry:

```
common --registry=https://raw.githubusercontent.com/filmil/bazel-registry/main
common --registry=https://bcr.bazel.build
```

The rules are documented in [docs/defs.md](docs/defs.md).

SKiDL runs on rules_python's hermetic CPython 3.12, with the locked wheels
of `requirements_lock.txt` and the KiCad 9.0.9 symbol libraries that
hdlf_rules_kicad pins (the SKiDL toolchain, issue #9). It needs no Python
on the host and no flag in your `.bazelrc`.
Under Bazel 8, rules_python asks for a C++ toolchain even though nothing
is compiled, so a build that finds none on the host must bring one, such
as hermetic_cc_toolchain's zig toolchain (`integration/MODULE.bazel` shows
how). Bazel 9 does not ask.

## Building

```sh
bazel build //...
bazel test //...
cd integration && bazel test //...
```

After changing a `.bzl` file, regenerate the documentation with
`bazel run //docs:update_docs`.
Format with `bazel run //tools:buildifier`.

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

This repository is under the Apache License 2.0 (`LICENSE`).
Prebuilt tools and libraries that the build downloads are listed here, with
their versions, sources and licenses, as they are added.

* SKiDL 3.0.0 (https://github.com/devbisme/skidl), kinet2pcb 1.1.4 and
  hierplace 1.1.0: MIT. PyPI has only their sdists; their wheels are built
  from those and mirrored in this repository's release
  `mirror-skidl-3.0.0` (`tools/mirror/README.md`).
* Their dependencies, wheels from PyPI pinned by hash in
  `requirements_lock.txt`: inspice 1.7.0.7 (GPL-3.0-or-later, which SKiDL
  needs on Python 3.11 and later); numpy, matplotlib, h5py, rich and the
  rest under BSD, MIT, PSF or similar licenses.
* The KiCad 9.0.9 symbol libraries, through hdlf_rules_kicad:
  CC-BY-SA-4.0 with the KiCad libraries exception.

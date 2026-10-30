<!-- SPDX-License-Identifier: Apache-2.0 -->
[![Test](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/test.yml)
[![Tag and Release](https://github.com/filmil/bazel_rules_netlist/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/releases)
[![Publish to my Bazel registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/bazel_rules_netlist/actions/workflows/publish-bcr.yml)

# rules_netlist

Hermetic Bazel rules for circuits written as code.
The build runs a SKiDL circuit script, writes a KiCad netlist that the
rules_kicad board rules use, and checks that a board's nets match it.
Every tool comes from a Bazel toolchain, so nothing needs to be installed
on the machine, and two builds give byte-identical outputs.

This repository is new. The rules arrive in milestones; see the
[issues](https://github.com/filmil/bazel_rules_netlist/issues) and the
[milestones](https://github.com/filmil/bazel_rules_netlist/milestones).
Only Linux x86_64 hosts are supported.

## Using it

```starlark
bazel_dep(name = "rules_netlist", version = "0.1.0")
```

That is for a release. rules_netlist builds on rules_kicad, which is in
no registry yet, and no release of rules_netlist is made until it is
(issue #8). To use rules_netlist meanwhile, pin a commit of it with an
`archive_override(module_name = "rules_netlist", ...)` of GitHub's archive
of that commit, and pin rules_kicad as rules_netlist does, in your
`MODULE.bazel` (an override counts only in the root module). Each is an
archive checked by its hash, so the build needs no `git` on the host:

```starlark
bazel_dep(name = "rules_kicad", version = "0.0.0")
archive_override(
    module_name = "rules_kicad",
    integrity = "sha256-L/dRAnTaCchSN66G81ijmGUiDzIUcU3DAy8RpNsaDMY=",
    strip_prefix = "bazel_rules_kicad-eda301dc59f9c29bd4fe2460bdd521eb9253a745",
    urls = ["https://github.com/filmil/bazel_rules_kicad/archive/eda301dc59f9c29bd4fe2460bdd521eb9253a745.tar.gz"],
)
```

`.bazelrc`, while the module is only in the filmil registry:

```
common --registry=https://raw.githubusercontent.com/filmil/bazel-registry/main
common --registry=https://bcr.bazel.build
```

The rules are documented in [docs/defs.md](docs/defs.md).

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

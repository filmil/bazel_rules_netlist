# SPDX-License-Identifier: Apache-2.0
"""Public API of hdlf_rules_netlist.

Load every rule from this file:

```starlark
load("@hdlf_rules_netlist//netlist:defs.bzl", "determinism_test", "swap_test", "tool_toolchain")
```

The rules arrive in milestone M6 (see the issues of
https://github.com/filmil/bazel_rules_netlist). This file is the only public
entry point; everything under `//netlist/private` may change without notice.
"""

load("//netlist/private:determinism.bzl", _determinism_test = "determinism_test")
load("//netlist/private:skidl_toolchain.bzl", _SkidlToolchainInfo = "SkidlToolchainInfo", _skidl_toolchain = "skidl_toolchain")
load("//netlist/private:swap_test.bzl", _swap_test = "swap_test")
load("//netlist/private:toolchains.bzl", _tool_toolchain = "tool_toolchain")

SkidlToolchainInfo = _SkidlToolchainInfo
determinism_test = _determinism_test
skidl_toolchain = _skidl_toolchain
swap_test = _swap_test
tool_toolchain = _tool_toolchain

# The release this file belongs to. It lets a consumer check, in a test,
# that the module it resolved is the one it expects.
RULES_NETLIST_API_VERSION = "0"

# SPDX-License-Identifier: Apache-2.0
"""Public API of rules_netlist.

Load every rule from this file. For now it holds only the API version:

```starlark
load("@rules_netlist//netlist:defs.bzl", "RULES_NETLIST_API_VERSION")
```

The rules arrive in milestone M6 (see the issues of
https://github.com/filmil/bazel_rules_netlist). This file is the only public
entry point; everything under `//netlist/private` may change without notice.
"""

# The release this file belongs to. It lets a consumer check, in a test,
# that the module it resolved is the one it expects.
RULES_NETLIST_API_VERSION = "0"

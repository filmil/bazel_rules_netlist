<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of rules_netlist.

Load every rule from this file. For now it holds only the API version:

```starlark
load("@rules_netlist//netlist:defs.bzl", "RULES_NETLIST_API_VERSION")
```

The rules arrive in milestone M6 (see the issues of
https://github.com/filmil/bazel_rules_netlist). This file is the only public
entry point; everything under `//netlist/private` may change without notice.


# SPDX-License-Identifier: Apache-2.0
"""The SKiDL toolchain's smoke test circuit: one resistor from the pinned
Device library. Writes what SKiDL found to the file named by argv[1]."""

import sys

from skidl import Part, get_default_tool

r = Part("Device", "R", value="1k", footprint="Resistor_SMD:R_0603_1608Metric")
with open(sys.argv[1], "w") as out:
    out.write("tool=%s\n" % get_default_tool())
    out.write("%s %s pins=%d\n" % (r.name, r.ref, len(r.pins)))
    for p in sorted(r.pins, key=lambda p: p.num):
        out.write("pin %s %s\n" % (p.num, p.name))

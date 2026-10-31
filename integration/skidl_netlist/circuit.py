# SPDX-License-Identifier: Apache-2.0
"""A LED and its resistor on a 3.3 V connector: the same circuit as
hdlf_rules_kicad's e2e design, as SKiDL code."""

from parts import led, resistor
from skidl import Net, Part

vcc = Net("+3V3")
gnd = Net("GND")

j1 = Part(
    "Connector_Generic",
    "Conn_01x02",
    ref="J1",
    value="Power",
    footprint="Connector_PinHeader_2.54mm:PinHeader_1x02_P2.54mm_Vertical",
)
r1 = resistor("R1", "1k")
d1 = led("D1")

vcc += j1[1], r1[1]
gnd += j1[2], d1["K"]
r1[2] += d1["A"]

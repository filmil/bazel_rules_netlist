# SPDX-License-Identifier: Apache-2.0
"""Parts for circuit.py, in another directory: skidl_netlist's deps."""

from skidl import Part


def resistor(ref, value):
    return Part("Device", "R", ref=ref, value=value, footprint="Resistor_SMD:R_0603_1608Metric")


def led(ref):
    return Part("Device", "LED", ref=ref, value="LED", footprint="LED_SMD:LED_0603_1608Metric")

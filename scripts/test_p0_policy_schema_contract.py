#!/usr/bin/env python3
"""Bounded P0 contract regression for the public policy JSON Schema.

No privileged operations, device access or network access. This independently
checks a small documented overlap with Go Bundle.Validate; it does not prove
general schema/Go equivalence or any live emergency-route reachability.
Requires python3-jsonschema (Draft 2020-12).
"""

from __future__ import annotations

import copy
import json
from pathlib import Path

from jsonschema import Draft202012Validator

ROOT = Path(__file__).resolve().parents[1]


def main() -> None:
    schema = json.loads((ROOT / "schemas/policy-v1alpha1.schema.json").read_text())
    baseline = json.loads((ROOT / "testdata/host-baseline.json").read_text())
    Draft202012Validator.check_schema(schema)
    validator = Draft202012Validator(schema)

    scenarios = (
        ("original baseline", "drop", True, ["192.0.2.0/24"], True),
        ("accept without rescue", "accept", False, [], True),
        ("accept with rescue", "accept", True, ["192.0.2.0/24"], True),
        ("drop with rescue", "drop", True, ["192.0.2.0/24"], True),
        ("reject with rescue", "reject", True, ["192.0.2.0/24"], True),
        ("drop disabled rescue", "drop", False, ["192.0.2.0/24"], False),
        ("reject disabled rescue", "reject", False, ["192.0.2.0/24"], False),
        ("drop empty rescue", "drop", True, [], False),
        ("reject empty rescue", "reject", True, [], False),
        ("drop no rescue", "drop", False, [], False),
        ("reject no rescue", "reject", False, [], False),
    )
    for name, default, enabled, networks, expected in scenarios:
        bundle = copy.deepcopy(baseline)
        bundle["spec"]["defaultInbound"] = default
        bundle["spec"]["emergencyAccess"] = {
            "enabled": enabled, "networks": networks
        }
        errors = list(validator.iter_errors(bundle))
        if (not errors) != expected:
            raise AssertionError(
                f"{name}: valid={not errors}, expected={expected}, errors={errors}"
            )

    port_scenarios = (
        ("TCP wildcard port", "tcp", 0, True),
        ("UDP wildcard port", "udp", 0, True),
        ("TCP upper bound", "tcp", 65535, True),
        ("UDP DNS port", "udp", 53, True),
        ("TCP negative port", "tcp", -1, False),
        ("UDP overflow port", "udp", 65536, False),
        ("ICMP zero port", "icmp", 0, True),
        ("ICMPv6 zero port", "icmpv6", 0, True),
        ("ICMP nonzero port", "icmp", 443, False),
        ("ICMPv6 nonzero port", "icmpv6", 53, False),
        ("TCP omitted port", "tcp", None, True),
        ("ICMP omitted port", "icmp", None, True),
    )
    for name, protocol, port, expected in port_scenarios:
        bundle = copy.deepcopy(baseline)
        rule = bundle["spec"]["rules"][0]
        rule["protocol"] = protocol
        if port is None:
            rule.pop("destinationPort", None)
        else:
            rule["destinationPort"] = port
        errors = list(validator.iter_errors(bundle))
        if (not errors) != expected:
            raise AssertionError(
                f"{name}: valid={not errors}, expected={expected}, errors={errors}"
            )

    for index in (0, 1):  # reserved wildcard must not become any alias type
        invalid = copy.deepcopy(baseline)
        invalid["spec"]["aliases"][index]["id"] = "any"
        if not list(validator.iter_errors(invalid)):
            raise AssertionError(f"reserved alias id any accepted for index={index}")

    print(
        "P0_POLICY_SCHEMA_CONTRACT_PASS "
        f"({len(scenarios)} rescue cases + {len(port_scenarios)} port cases "
        "+ 2 reserved-alias negatives)"
    )


if __name__ == "__main__":
    main()

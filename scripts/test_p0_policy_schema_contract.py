#!/usr/bin/env python3
"""Bounded P0 policy-schema regression, with optional real Go CLI parity.

Only documented overlapping semantics are tested; do NOT infer full JSON
Schema vs Go equivalence, owner authority, rescue reachability or kernel safety.
Run `python3 scripts/test_p0_policy_schema_contract.py --verify-cli` to also
build the unprivileged lunactl binary and validate every synthetic case through
Go policy.Decode. Requires jsonschema and an offline-available Go toolchain.
"""

from __future__ import annotations

import argparse
import copy
import json
import os
import subprocess
import tempfile
from pathlib import Path

from jsonschema import Draft202012Validator

ROOT = Path(__file__).resolve().parents[1]


def verify_contract(verify_cli: bool) -> None:
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
    note_scenarios = (
        ("ordinary note", "Safe note", True),
        ("empty note", "", True),
        ("linefeed note", "two\nlines", True),
        ("NUL in note", "before\x00after", False),
        ("NUL only", "\x00", False),
        ("maximum ASCII note", "x" * 512, True),
        ("oversized ASCII note", "x" * 513, False),
    )
    cases: list[tuple[str, dict, bool]] = []
    for name, default, enabled, networks, expected in scenarios:
        bundle = copy.deepcopy(baseline)
        bundle["spec"]["defaultInbound"] = default
        bundle["spec"]["emergencyAccess"] = {
            "enabled": enabled, "networks": networks
        }
        cases.append((f"rescue/{name}", bundle, expected))

    for name, protocol, port, expected in port_scenarios:
        bundle = copy.deepcopy(baseline)
        rule = bundle["spec"]["rules"][0]
        rule["protocol"] = protocol
        if port is None:
            rule.pop("destinationPort", None)
        else:
            rule["destinationPort"] = port
        cases.append((f"port/{name}", bundle, expected))

    for index in (0, 1):
        invalid = copy.deepcopy(baseline)
        invalid["spec"]["aliases"][index]["id"] = "any"
        cases.append((f"reserved alias/{index}", invalid, False))

    for name, note, expected in note_scenarios:
        bundle = copy.deepcopy(baseline)
        bundle["spec"]["rules"][0]["note"] = note
        cases.append((f"note/{name}", bundle, expected))

    # Build once into a temporary directory; all policy files are synthetic and
    # discarded. Running `validate` is non-enforcing and never touches nft.
    with tempfile.TemporaryDirectory(prefix="lightluna-schema-contract-") as tmp:
        temporary = Path(tmp)
        cli = temporary / "lunactl"
        fixture = temporary / "policy.json"
        if verify_cli:
            env = dict(os.environ)
            env.update(GOPROXY="off", GOSUMDB="off", GOTOOLCHAIN="local")
            built = subprocess.run(
                ["go", "build", "-o", str(cli), "./cmd/lunactl"],
                cwd=ROOT, env=env, capture_output=True, text=True, timeout=90
            )
            if built.returncode:
                raise RuntimeError(f"offline lunactl build failed: {built.stderr}")

        for name, bundle, expected in cases:
            errors = list(validator.iter_errors(bundle))
            schema_valid = not errors
            if schema_valid != expected:
                raise AssertionError(
                    f"{name}: schema valid={schema_valid}, expected={expected}; "
                    f"errors={errors}"
                )
            if verify_cli:
                fixture.write_text(
                    json.dumps(bundle, ensure_ascii=False), encoding="utf-8"
                )
                checked = subprocess.run(
                    [str(cli), "firewall", "policies", "validate", str(fixture)],
                    cwd=ROOT, capture_output=True, text=True, timeout=15
                )
                go_valid = checked.returncode == 0
                if go_valid != expected or go_valid != schema_valid:
                    raise AssertionError(
                        f"{name}: Go={go_valid}, schema={schema_valid}, "
                        f"expected={expected}; stderr={checked.stderr}"
                    )

    print(
        "P0_POLICY_SCHEMA_CONTRACT_PASS "
        f"({len(scenarios)} rescue + {len(port_scenarios)} port + "
        f"2 alias + {len(note_scenarios)} note = {len(cases)} cases; "
        f"Go CLI parity={'PASS' if verify_cli else 'NOT_RUN'})"
    )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--verify-cli", action="store_true",
        help="build unprivileged lunactl offline and validate all cases through Go"
    )
    arguments = parser.parse_args()
    verify_contract(arguments.verify_cli)

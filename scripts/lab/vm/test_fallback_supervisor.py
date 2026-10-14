#!/usr/bin/env python3
"""Offline P0 fixed-synthetic-table whitelist; never touches kernel."""
import importlib.util
import unittest
from pathlib import Path

spec=importlib.util.spec_from_file_location("vm_fallback_guard",Path(__file__).with_name("guest-fallback-supervisor.py"))
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)

def table():
    def match(spec):
        protocol,field,right=spec
        return {"match":{"op":"==","left":{"payload":{"protocol":protocol,"field":field}},"right":right}}
    rules=[]
    for pair in m.EXPECTED_MATCHES:
        rules.append({"rule":{"family":"inet","table":"luna_p0_stage","chain":"input",
                "expr":[match(pair[0]),match(pair[1]),
                        {"counter":{"packets":4,"bytes":300}},{"drop":None}]}})
    return {"nftables":[{"metainfo":{"version":"1.0"}},
      {"table":{"family":"inet","name":"luna_p0_stage","handle":1}},
      {"chain":{"family":"inet","table":"luna_p0_stage","name":"input","type":"filter","hook":"input","prio":-50,"policy":"accept"}},
      *rules]}

class GuardTests(unittest.TestCase):
    def test_positive_fixed_structure(self): self.assertTrue(m.fixed_synthetic_table(table()))
    def test_unexpected_extra_rule_denied(self):
        t=table();t["nftables"].append(t["nftables"][-1])
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_extra_chain_denied(self):
        t=table();t["nftables"].append({"chain":{"family":"inet","table":"luna_p0_stage","name":"foreign"}})
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_wrong_hook_denied(self):
        t=table();t["nftables"][2]["chain"]["hook"]="forward"
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_mismatched_table_denied(self):
        t=table();t["nftables"][1]["table"]["name"]="foreign"
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_non_drop_denied(self):
        t=table();t["nftables"][-1]["rule"]["expr"][-1]={"accept":None}
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_bad_json_denied(self):
        self.assertFalse(m.fixed_synthetic_table({"nftables":{}}))
    def test_wrong_tcp_port_denied(self):
        t=table(); t["nftables"][3]["rule"]["expr"][1]["match"]["right"]=22
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_duplicate_rules_not_allowed(self):
        t=table(); t["nftables"][4]=t["nftables"][3]
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_extra_match_not_allowed(self):
        t=table(); t["nftables"][3]["rule"]["expr"].insert(2,t["nftables"][3]["rule"]["expr"][1])
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_wrong_ipv6_source_denied(self):
        t=table(); t["nftables"][5]["rule"]["expr"][0]["match"]["right"]="::1"
        self.assertFalse(m.fixed_synthetic_table(t))
    def test_wrong_rule_chain_denied(self):
        t=table();t["nftables"][-1]["rule"]["chain"]="other"
        self.assertFalse(m.fixed_synthetic_table(t))

if __name__=="__main__":unittest.main(verbosity=2)

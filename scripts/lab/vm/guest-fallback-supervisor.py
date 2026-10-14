#!/usr/bin/env python3
"""P0 VM-only SECOND PID1-managed synthetic namespace watchdog prototype.

Runs from systemd under nsenter -n BEFORE the synthetic nft hook is applied:
the independent process itself pins the target guest netns, even if keeper dies.
Only a fixed synthetic table may be removed; never product nft policy.
"""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

ROOT = Path("/opt/luna-p0-kernel-tests")
READY = ROOT / "fallback-ready.json"
DONE = ROOT / "fallback-done.json"
REJECTED = ROOT / "fallback-rejected.json"
TABLE = "luna_p0_stage"
STAGE = ROOT / "guest-stage-journal.py"
EXPECTED_MATCHES = (
    (("ip", "saddr", "198.19.0.2"), ("tcp", "dport", 18080)),
    (("ip", "saddr", "198.19.0.2"), ("udp", "dport", 18081)),
    (("ip6", "saddr", "fd42:198:19::2"), ("icmpv6", "type", "echo-request")),
)


def persist(path, value):
    fd, name = tempfile.mkstemp(prefix=".fallback-", dir=ROOT)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            json.dump(value, f, sort_keys=True, separators=(",", ":"))
            f.write("\n")
            f.flush()
            os.fsync(f.fileno())
        if path.exists() or path.is_symlink():
            raise ValueError("fallback evidence file already exists")
        os.replace(name, path)
        fd_dir = os.open(ROOT, os.O_RDONLY | os.O_DIRECTORY)
        try: os.fsync(fd_dir)
        finally: os.close(fd_dir)
    finally:
        if os.path.exists(name): os.unlink(name)


def cmd(*args):
    return subprocess.run(args, text=True, capture_output=True, check=False, timeout=4)


def module():
    if STAGE.is_symlink() or not STAGE.is_file():
        raise ValueError("missing or symlinked stage module")
    spec = importlib.util.spec_from_file_location("vm_fallback_journal", STAGE)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def pid_ticks(pid):
    return int(Path("/proc/"+str(pid)+"/stat").read_text().rsplit(") ",1)[1].split()[19])


def service_dead(unit):
    if not (unit.startswith("luna-p0-auto-wd-") or unit.startswith("luna-p0-stage-keeper-")) or not unit.endswith(".service"):
        raise ValueError("unexpected systemd unit")
    p=cmd("systemctl","show",unit,"-p","MainPID","--value")
    if p.returncode!=0 or not p.stdout.strip().isdigit():
        raise ValueError("cannot check systemd MainPID")
    return int(p.stdout.strip())==0


def fixed_synthetic_table(payload):
    """Require exact lab table/chain and exactly three drop rules, no foreign chains."""
    if not isinstance(payload,dict) or set(payload)!={"nftables"} or not isinstance(payload["nftables"],list):
        return False
    entries=[e for e in payload["nftables"] if isinstance(e,dict) and "metainfo" not in e]
    tables=[e["table"] for e in entries if set(e)=={"table"}]
    chains=[e["chain"] for e in entries if set(e)=={"chain"}]
    rules=[e["rule"] for e in entries if set(e)=={"rule"}]
    if len(entries)!=5 or len(tables)!=1 or len(chains)!=1 or len(rules)!=3:
        return False
    if any(t.get("family")!="inet" or t.get("name")!=TABLE for t in tables):
        return False
    ch=chains[0]
    if (ch.get("family")!="inet" or ch.get("table")!=TABLE or ch.get("name")!="input"
        or ch.get("type")!="filter" or ch.get("hook")!="input"
        or ch.get("prio")!=-50 or ch.get("policy")!="accept"):
        return False
    def expected_match(spec):
        protocol, field, right = spec
        return {"match":{"op":"==","left":{"payload":{"protocol":protocol,"field":field}},"right":right}}
    for r, fixed in zip(rules, EXPECTED_MATCHES):
        if r.get("family")!="inet" or r.get("table")!=TABLE or r.get("chain")!="input":
            return False
        if set(r)-{"family","table","chain","handle","expr"}:
            return False
        expr=r.get("expr")
        if not isinstance(expr,list) or len(expr)!=4:
            return False
        if expr[:2]!=[expected_match(fixed[0]),expected_match(fixed[1])]:
            return False
        if expr[-1]!={"drop":None}:
            return False
        if not isinstance(expr[2],dict) or set(expr[2])!={"counter"}:
            return False
        counters=expr[2]["counter"]
        if (not isinstance(counters,dict) or set(counters)!={"packets","bytes"}
            or any(type(v) is not int or v<0 for v in counters.values())):
            return False
    return True


def main():
    if os.geteuid()!=0 or not Path("/etc/light-luna-p0-kernel-vm").is_file():
        raise ValueError("root VM marker required")
    if sorted(os.listdir("/sys/class/net"))!=["lo"] or cmd("systemd-detect-virt").stdout.strip()!="kvm":
        raise ValueError("only NIC-less KVM VM allowed")
    if len(sys.argv) not in (6,7):
        raise ValueError("usage keeperPID keeperTicks targetNetns mainUnit keeperUnit [--lab-fault-after-delete|--lab-race-at-deadline]")
    crash_after_delete = len(sys.argv)==7 and sys.argv[6]=="--lab-fault-after-delete"
    race_at_deadline = len(sys.argv)==7 and sys.argv[6]=="--lab-race-at-deadline"
    if len(sys.argv)==7 and not (crash_after_delete or race_at_deadline):
        raise ValueError("unknown VM-only fault mode")
    if crash_after_delete and os.environ.get("LUNA_VM_FALLBACK_CRASH_ACK") != "VM_ONLY_CRASH_AFTER_SYNTHETIC_DELETE":
        raise ValueError("VM-only crash failpoint not explicitly acknowledged")
    if race_at_deadline and os.environ.get("LUNA_VM_FALLBACK_RACE_ACK") != "VM_ONLY_PRIMARY_FALLBACK_SAME_DEADLINE":
        raise ValueError("simultaneous-deadline lab check not acknowledged")
    keeper=int(sys.argv[1])
    ticks=int(sys.argv[2])
    netns=sys.argv[3]
    main_unit=sys.argv[4]
    keeper_unit=sys.argv[5]
    if not netns.startswith("net:[") or not netns.endswith("]"):
        raise ValueError("invalid target network namespace")
    if keeper<=1 or pid_ticks(keeper)!=ticks or os.readlink("/proc/"+str(keeper)+"/ns/net")!=netns:
        raise ValueError("keeper PID/start identity mismatched before arm")
    # Because this systemd process itself is started via nsenter -n, it
    # remains in the target namespace even when the keeper is SIGKILLed.
    if os.readlink("/proc/self/ns/net")!=netns:
        raise ValueError("independent fallback is not inside expected netns")
    if READY.exists() or DONE.exists() or REJECTED.exists():
        raise ValueError("stale fallback records")
    mod=module()
    plan=mod.manifest()
    journal=mod.safe_load(mod.JOURNAL)
    boot=mod.boot_id()
    if journal.get("bootId")!=boot or journal.get("stageId")!=plan["stageId"] or journal.get("state")!="VALIDATED":
        raise ValueError("fallback must be armed before synthetic commit")
    primary=mod.safe_load(ROOT/"auto-watchdog-ready.json")
    if (primary.get("bootId")!=boot or primary.get("stageId")!=plan["stageId"]
        or primary.get("netnsInode")!=netns or primary.get("anchorPid")!=keeper
        or primary.get("anchorBirthTicks")!=ticks):
        raise ValueError("primary watchdog READY is not matched")
    before=cmd("nft","list","table","inet",TABLE)
    if before.returncode==0: raise ValueError("synthetic table already exists before arm")
    deadline=primary["deadlineMonotonic"]
    if type(deadline) not in (float,int) or deadline-time.monotonic()<5:
        raise ValueError("too close to deadline")
    persist(READY,{"schema":"lightluna.vm-fallback-ready/v1alpha1","stageId":plan["stageId"],
                   "bootId":boot,"netnsInode":netns,"keeperBirthTicks":ticks,
                   "deadlineMonotonic":deadline,"primaryUnit":main_unit,"keeperUnit":keeper_unit,
                   "kernelWrites":False,"ownerApproved":False})
    print("VM_FALLBACK_PRECOMMIT_ARMED_PASS",flush=True)
    remaining=deadline+(0.0 if race_at_deadline else 1.0)-time.monotonic()
    if remaining>0: time.sleep(remaining)
    if mod.boot_id()!=boot or os.readlink("/proc/self/ns/net")!=netns:
        raise ValueError("boot/netns identity changed")
    fence_path=ROOT/"guest-rollback-fence.py"
    if fence_path.is_symlink() or not fence_path.is_file():
        raise ValueError("missing guarded VM rollback fence")
    spec=importlib.util.spec_from_file_location("vm_fallback_rollback_fence",fence_path)
    fence=importlib.util.module_from_spec(spec)
    spec.loader.exec_module(fence)
    with fence.guard(mod,plan["stageId"],boot) as current:
        # All decisions below are based on the journal AFTER the exclusive
        # cross-process VM lock. The primary uses the identical lock path.
        if os.readlink("/proc/self/ns/net")!=netns or mod.boot_id()!=boot:
            raise ValueError("fallback lost verified namespace/boot under lock")
        if current["state"]=="ROLLED_BACK":
            persist(REJECTED,{"status":"PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER","stageId":plan["stageId"],"rollbackFenced":True})
            return
        if current["state"]!="COMMITTED_SYNTHETIC":
            persist(REJECTED,{"status":"NOT_COMMITTED_NO_SUCCESS","stageId":plan["stageId"],"rollbackFenced":True})
            return
        if not (service_dead(main_unit) or service_dead(keeper_unit)):
            persist(REJECTED,{"status":"PRIMARY_AND_KEEPER_STILL_ACTIVE_REFUSED_TAKEOVER","stageId":plan["stageId"],"rollbackFenced":True})
            return
        snapshot=cmd("nft","-j","list","table","inet",TABLE)
        if snapshot.returncode!=0 or not fixed_synthetic_table(json.loads(snapshot.stdout)):
            raise ValueError("synthetic table shape invalid under fence; refuse deletion")
        # ONLY the fixed test table in the already-pinned guest namespace.
        removed=cmd("nft","delete","table","inet",TABLE)
        if removed.returncode!=0 or cmd("nft","list","table","inet",TABLE).returncode==0:
            raise ValueError("fixed synthetic table rollback could not be verified")
        # This still is NOT atomic across kernel and journal disk fsync;
        # process termination drops flock and must NOT imply success.
        if crash_after_delete:
            print("VM_FALLBACK_FAILPOINT_AFTER_DELETE_BEFORE_JOURNAL", flush=True)
            os._exit(72)
        mod.transition("transition","ROLLED_BACK",plan)
        persist(DONE,{"schema":"lightluna.vm-fallback-done/v1alpha1",
                      "status":"SYNTHETIC_FALLBACK_ROLLBACK_CONFIRMED",
                      "stageId":plan["stageId"],"bootId":boot,"netnsInode":netns,
                      "deadlineMonotonic":deadline,"completedMonotonic":time.monotonic(),
                      "ownerApproved":False,"syntheticGuestOnly":True,"rollbackFenced":True})
    print("VM_FALLBACK_INDEPENDENT_PID1_SYNTHETIC_ROLLBACK_PASS",flush=True)


if __name__=="__main__":
    try: main()
    except Exception as exc:
        print("VM_FALLBACK_FAILED:",type(exc).__name__,str(exc)[:200],file=sys.stderr,flush=True)
        sys.exit(3)

#!/usr/bin/env python3
"""P0 SYNTHETIC VM-only autonomous watchdog. NEVER production execution.

Runs under VM PID1 as a separate systemd service. It receives an anchor
PID + /proc process birth time + netns inode and can touch ONLY a fixed
synthetic test table in a KVM VM with no external NIC. It does not use the
Light Luna nft compiler script, and it does not authenticate an owner.
"""
import importlib.util
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import time

ROOT = Path("/opt/luna-p0-kernel-tests")
STAGE_MODULE = ROOT / "guest-stage-journal.py"
READY = ROOT / "auto-watchdog-ready.json"
DONE = ROOT / "auto-watchdog-done.json"
TABLE = "luna_p0_stage"

def save(path, object_):
    fd,name = tempfile.mkstemp(prefix=".auto-wd-",dir=ROOT)
    try:
        os.fchmod(fd,0o600)
        with os.fdopen(fd,"w",encoding="utf-8") as f:
            json.dump(object_,f,sort_keys=True,separators=(",",":"))
            f.write("\n")
            f.flush()
            os.fsync(f.fileno())
        os.replace(name,path)
        directory = os.open(ROOT,os.O_RDONLY | os.O_DIRECTORY)
        try: os.fsync(directory)
        finally: os.close(directory)
    finally:
        if os.path.exists(name): os.unlink(name)

def birth(pid):
    # process stat field 22 (starttime); split after ") " for comm with spaces
    fields=Path(f"/proc/{pid}/stat").read_text().rsplit(") ",1)[1].split()
    return int(fields[19])

def namespace(pid):
    return os.readlink(f"/proc/{pid}/ns/net")

def run_nft(pid,*args,expect_success=True):
    result=subprocess.run(["nsenter","-t",str(pid),"-n","nft",*args],
                          text=True,capture_output=True,timeout=3,check=False)
    if expect_success and result.returncode!=0:
        raise ValueError(f"nft operation denied or failed rc={result.returncode}: {result.stderr[:200]}")
    return result

def main():
    if os.geteuid()!=0 or not Path("/etc/light-luna-p0-kernel-vm").is_file():
        raise ValueError("not authorized VM test root")
    if os.listdir("/sys/class/net")!=["lo"] or subprocess.check_output(
        ["systemd-detect-virt"],text=True).strip()!="kvm":
        raise ValueError("VM has network NIC or not KVM")
    if len(sys.argv)!=4: raise ValueError("usage: watchdog.py anchorPID start_ticks net_inode")
    pid=int(sys.argv[1])
    ticks=int(sys.argv[2])
    inode=sys.argv[3]
    if not re.fullmatch(r"net:\[[0-9]+\]",inode): raise ValueError("invalid netns identity")
    if pid<=1 or birth(pid)!=ticks or namespace(pid)!=inode:
        raise ValueError("PID identity or network namespace mismatch before arm")
    spec=importlib.util.spec_from_file_location("vm_stage_journal",STAGE_MODULE)
    mod=importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    manifest=mod.manifest()
    journal=mod.safe_load(mod.JOURNAL)
    if journal["stageId"]!=manifest["stageId"] or journal["state"]!="VALIDATED":
        raise ValueError("journal was not VALIDATED before watchdog arming")
    if READY.exists() or DONE.exists():
        raise ValueError("stale ready/done watchdog records")
    if run_nft(pid,"list","table","inet",TABLE,expect_success=False).returncode==0:
        raise ValueError("synthetic table exists BEFORE watchdog is armed")
    ttl=manifest["ttlSeconds"]
    armed=time.monotonic()
    deadline=armed+ttl
    save(READY,{"stageId":manifest["stageId"],"anchorPid":pid,
                "anchorBirthTicks":ticks,"netnsInode":inode,
                "armedMonotonic":armed,"deadlineMonotonic":deadline,
                "ttlSeconds":ttl,"armed":True})
    print("VM_AUTO_WATCHDOG_ARMED_READY",flush=True)
    while True:
        remain=deadline-time.monotonic()
        if remain<=0:break
        time.sleep(min(0.2,remain))
    # All kernel operations are limited to an already-verified guest-only netns.
    if birth(pid)!=ticks or namespace(pid)!=inode:
        raise ValueError("anchor identity no longer trusted at rollback")
    latest=mod.safe_load(mod.JOURNAL)
    if latest["stageId"]!=manifest["stageId"]:
        raise ValueError("transaction changed while watchdog armed")
    existing=run_nft(pid,"list","table","inet",TABLE,expect_success=False)
    if latest["state"]!="COMMITTED_SYNTHETIC":
        if existing.returncode==0:
            run_nft(pid,"delete","table","inet",TABLE)
        raise ValueError("not committed at TTL; fail-closed, no success receipt")
    if existing.returncode!=0:
        raise ValueError("table already missing before automatic rollback, refusing unverifiable success")
    run_nft(pid,"delete","table","inet",TABLE)
    if run_nft(pid,"list","table","inet",TABLE,expect_success=False).returncode==0:
        raise ValueError("synthetic table still exists after deletion")
    mod.transition("transition","ROLLED_BACK",manifest)
    completed=time.monotonic()
    save(DONE,{"stageId":manifest["stageId"],"status":"AUTO_ROLLBACK_CONFIRMED",
               "completedMonotonic":completed,"deadlineMonotonic":deadline,
               "lagSeconds":max(0.0,completed-deadline)})
    print("VM_AUTO_WATCHDOG_SELF_TRIGGERED_ROLLBACK_PASS",flush=True)

if __name__=="__main__":
    try: main()
    except Exception as exc:
        print("VM_AUTO_WATCHDOG_FAILED:",repr(exc),file=sys.stderr,flush=True)
        sys.exit(3)

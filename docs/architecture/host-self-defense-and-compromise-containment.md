# Self-Defense, Host Threat Detection and Cloud/Peer Containment

Status: **DECIDED architectural contract from ADR-028, NOT IMPLEMENTED**
Date: 2026-10-14
Scope: Linux host integrity, process/file/network runtime signals, optional malware file scanner, local safety and Cloud quarantine of compromised-node intelligence
Primary sources: [ADR-028](../adr/ADR-028-self-defense-host-compromise.md), [Security](../07-SECURITY-SAFETY.md), [eBPF/XDP boundaries](../adr/ADR-021-ebpf-observability-xdp-optional.md), [Collective Immune Memory](federated-immune-threat-intelligence.md)
Schema draft: [HostIntegrityReport](../../schemas/host-integrity-report-v1alpha1.schema.json); [fixture](../../testdata/host-integrity-report.synthetic.json)
Source references: [Tetragon](https://tetragon.io/docs/overview/), [Falco](https://falco.org/docs/), [IMA](https://github.com/linux-integrity/ima-doc/blob/main/ima-concepts.rst), [fs-verity](https://www.kernel.org/doc/html/latest/filesystems/fsverity.html), [ClamAV](https://docs.clamav.net/manual/OnAccess.html)

## 1. Security objective and assurance ceiling

Three distinct issues:
1. A malicious process on a Linux host may attack services, encrypt files, launch bot/C2 traffic, spread laterally or target Light Luna configuration.
2. Light Luna itself, privileged nftables/BPF loader, signed updater, or credentials could be subverted, disabling local defenses and altering what Cloud sees.
3. A compromised but authenticated node can intentionally send forged reports, contaminated scenarios or exploits disguised as samples. Without Cloud ingestion isolation it might poison shared knowledge and spread harms to other clients.

**The objective is bounded detection, containment and recovery, not guaranteed virus-free attestation.** A root/kernel attacker can modify or suppress in-host telemetry, alter BPF programs and seize local software credentials. A positive on-host scan cannot establish that the kernel is uncompromised. Off-host telemetry and hardware-based attestation may increase assurance on supported machines but still require sound provisioning/verifiers.

## 2. High-level trust topology

```text
Host app/process/files/network activity
     |         |                    |
 eBPF tracing  journald/proc   optional file scanner (ClamAV/fanotify)
(optional)     hash/policy     bounded, isolated on-demand default
     +---------+--------------------+
               |
          Go Host Observer (nonroot/limited cap)
               |
     +---------+---------------+
     |                         |
  Independent Self-Integrity   Go Luna Brain (hypotheses)
  Guard: signed file/config    evidence graph, benign counterexamples
  digests, process health,     and risk proposals only
  BPF attach ownership        |
     +-------------+-----------+
                   |
            Independent Policy Gate
             scoped TTL, local owner, rollback
                   |
       WARN / ESCALATE / SCOPED ISOLATION
                   |
                Host Health Receipt
                   |
     optional privacy-minimized attested upload
                   |
      Cloud Device Security Ingress (UNTRUSTED)
       strict schema, auth, limits, dedup, quarantine
                   |
     +-----------------------------+
     | compromised/suspect trust   |
     | suspend publishing authority|
     | no global trust contribution|
     +-----------------------------+
                   |
         Separate clean-room review
           approved signed intelligence
                   |
        Other clients receive data-only
          watch / revoked-identity notices
        no raw malware, no executable content
```

The host defensive guardian and user-facing Brain use different privilege and fault domains; neither can rewrite the signing trust roots, its own permissions or independent firewall watchdog. Cloud agents use outbound network only; neither remote Cloud identity nor eBPF telemetry grants runtime kernel policy writes.

## 3. What each tool can actually observe

| Technology | Useful signal | Limitation / rollout |
|---|---|---|
| XDP | early ingress packet-rate/malformed packet/counters, supported L3/L4 signatures | no files/process memory, cannot verify malware or decrypt TLS, no automatic packet-drop until separate gate |
| eBPF tracepoint/kprobe/LSM (optional) | process exec/fork ancestry, credential/capability change, TCP connect, selected file events, namespace or sensitive syscall use | event blind spots/kernel differences, attacker with kernel privileges may tamper; start observe-only |
| journald/audit/proc/packaging metadata | service restarts, persistence/service changes, package/exec hashes, process/cgroup names, auth failures | not always independent, race/coverage gaps, attacker can alter if privileged |
| file integrity checks / IMA / fs-verity where supported | signed binaries/config integrity mismatch, unauthorized module/agent changes | not anti-malware classification by itself; enable only where compatible; trusted boot/TPM optional |
| optional AV signatures + fanotify | malware pattern detection on selected files; on-access protection possible if separately enabled | signature blind spots, performance and lockout risks, root recursive scan can deadlock, no clean verdict from one scan |
| network egress+DNS | unexpected C2-like callbacks, beacon timing, rare destinations, lateral port fan-out | DNS over HTTPS and TLS conceal some content; legitimate update/backup/tools cause false positives |
| service impact and cross-sensor correlation | process + network + file change together improves triage | correlated sensors aren't independent ground truth |

A V1 'Quick Scan' should be **non-destructive**: inventory running services/ports, unusual persistence mechanisms, executable hashes against known-good signed release, Linux auth events, scan selected high-risk directories (when AV module installed), and network connection anomalies. Output lists findings and uncovered surfaces, not `NO MALWARE`. An uninstalled file scanner shows `NOT_AVAILABLE`, never `CLEAN`.

## 4. Host Guardian lifecycle and bounded actions

States:
- `HEALTHY_OBSERVED`: supported checks completed with valid signatures, good freshness, no observed anomaly; explicitly **not proof of clean host**.
- `UNKNOWN`: telemetry missing/stale, unsupported kernel hooks, or incomplete file coverage.
- `SUSPECTED`: suspicious behavior, contradictory signals, integrity drift.
- `CONTAINED`: explicit owner-approved local containment verified for named process/service/egress scope; not equivalent to eradicating infection.
- `COMPROMISED_SUSPECTED`: critical tamper, verified malicious artifact/command-and-control evidence or corroborated compromise case; response escalation and restricted Cloud contribution.
- `RECOVERY_REQUIRED`: local trust basis irreparably lost or hard integrity failure; manual trusted rescue and key rotation required.

`UNKNOWN` and `COMPROMISED_SUSPECTED` must not be collapsed into one "infected" label or auto-reinstall event. Host results include evidence IDs, collection coverage/loss, last signature freshness, signed file hashes, relevant local system versions, independent corroboration, precision unknowns and suggested *non-destructive next step*.

Severity-based mapping (proposed, not deployed):

| Finding | Default action | Owner-gated escalation |
|---|---|---|
| low-confidence beacon-like DNS | observe + explain + check benign alternates | more local observation, TTL-based domain policy proposal |
| repeated lateral SMB/SSH scan from local process | alert, process/cgroup scope and timeline | bounded egress containment with protected management paths |
| signed Light Luna binary/config mismatch | stop unsafe software updates/remote-sensitive actions, integrity alarm | revoke client contribution rights and trusted rescue |
| suspicious mass file encryption/rename | urgent alert and measured scope, avoid destroying forensic evidence | optional scoped process pause/egress restriction only in tested modes |
| confirmed malicious file by optional scanner | local finding, isolate sample in protected *local* quarantine subject to scanner safety | restore file with approval; never ship file to Cloud by default |
| suspected root/kernel compromise | mark untrusted, revoke Cloud authority externally, stop claiming healthy | external console/rescue/reimage/rotate keys; in-host self-heal untrusted |

**Never auto kill PID 1, systemd, sshd, Docker/NetBird or the only management link based on one heuristic.** Process identity may be recycled; use stable process handle/cgroup evidence with lifetime safeguards. egress-block should account for Cloud control/backup/owner path, but preserve independent physical recovery even if Cloud is itself hostile.

## 5. Cloud contamination barrier

The cloud has **two parallel identity states**:
- cryptographic device identity authenticates *which enrolled key* sent data;
- host integrity attestation/reputation is **separate**, time-limited, scoped, and potentially unavailable.

Every inbound body is untrusted, including signed reports from `HEALTHY_OBSERVED` devices:
1. Terminate TLS in hardened API, verify device signature/tenant/revocation/nonce, apply size/time/field-cardinality limits and reject unknown binary/script fields.
2. Store only permitted typed/redacted metadata in segregated `UNTRUSTED_CANDIDATE` intake; no direct workflow execution, no dynamic SQL, no trust-root writes and no client-supplied executable update artifact.
3. Per-tenant source isolation, per-device and per-credential rate limits, offline/idempotency dedup, content validation and signature hash alignment; no raw logs/payload samples.
4. Evaluate integrity status with freshness, independent evidence and source quality. `SUSPECTED`, `RECOVERY_REQUIRED`, unknown attestation/explicit quarantine → **zero trusted consensus contribution**, but may allow tightly bounded incident alert/status paths; avoid auto-disabling basic device access based only on a spoofed signal.
5. Analysts/approved deterministic reviews may produce data-only scenario/behavior candidates. Critical threat bulletins need independent corroboration + benign false-positive testing; signing publisher separated from raw intake and model inference. Quarantine/retract a report/pack from compromised source; do not trust a former signed receipt as proof clean.
6. One node's `locallyBlocked` flag is *never* a Cloud global blacklist order. Other subscribers receive validated signed behavior knowledge and independently evaluate before action.
7. If client keys compromised: revoke device identity after authenticated owner/security process, expire pending remote config, rotate keys on a **clean trusted device**; a mere app reinstall using stolen original credential is not automatically a safe enrollment.

**Malware cannot normally "infect" Cloud simply by sending inert structured JSON**, but parsers, deserializers, file uploads, dynamic rules, model inputs, admin actions and artifact publishing are attack surfaces. Clean-room ingestion and a completely separate content-signing trust boundary are mandatory; avoid automatic Cloud-to-client execution by design.

## 6. Same-tenant lateral protection vs cross-tenant publicity

- **Same tenant:** owner can opt into immediate high-priority `SUSPECTED COMPROMISE` notices with minimal metadata. Fleet admin can view which device is suspected and revoke/suspend its remote management rights. Other local clients tighten *observation* of corresponding lateral behavior; extra blocking remains site-specific and locally approved.
- **Across tenants:** no automatic account/host identity disclosure. Publish only independently curated scenario patterns and scoped public indicators after consent, review and privacy checks. No Cloud operator can arbitrarily make all clients quarantine another account.
- **Management plane compromise:** lockdown Remote Console intents if cloud account taken over or trusted signer fails, keep read-only status if safe, verify on independent local owner path; suspect compromised Cloud code cannot push unsigned executables through knowledge channel.
- **Ransomware outbreak:** protect backup retention and offsite copies; do not let compromised client request Cloud delete/rekey or overwrite backup for all tenants; require strong owner approval for destructive Cloud actions.

## 7. Test/measurement gates

**P0 architectural schema and negative tests:** `HostIntegrityReportV1alpha1` with explicit `UNKNOWN`/limited coverage/authority; reject client attempts to claim guaranteed clean or issue Cloud-wide block/shell instructions. No live scans or eBPF attach.

**P1/P2 standalone host tests:** baseline permission isolation, eBPF *observe-only optional*, file hashing/auth log/process graph/egress correlation, bounded scanner tests using benign synthetic/EICAR **only in disposable VM**, recovery/privilege drops, 30/90-day telemetry retention, kernel arch detection. Avoid running live malware, mass-encryption triggers or invasive scanner on HOMESV/production.

**P3 isolated containment:** under approved lab, test process-specific egress limits, false positives, admin recovery/SSH continuity, rollback after crash/CPU/disk full, anti-lockout, PID reuse and Cloud-disconnected state. Only after independent Policy Gate and owner approval.

**P7 Cloud trust:** authenticated compromised-client simulation with forged signals, non-executable malicious-looking JSON, poisoned scenario/hash, remote access takeover, cloned identity/Sybil, contaminated models, cross-tenant data exfiltration and revoked device; verify Cloud quarantine and pack publisher isolation. Do not measure `malware detection 100%`; report measured coverage for supported threat classes and unsupported vectors.

No working Host Guardian, scanner, endpoint EDR, attestation, Cloud quarantine engine or privileged BPF loader exists in current code. All described commands/state transitions are future engineering contracts.

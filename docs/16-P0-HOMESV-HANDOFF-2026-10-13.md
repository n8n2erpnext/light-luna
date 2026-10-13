# 16 — HOMESV P0 Engineering Handoff (2026-10-13)

## Truthful state

- `main` on VPS-ARM and GitHub contains the 19 original docs at `66920c251603b0b0c4bcd0d0cfc94ca8f7220d41` (parent `e6daaa7`).
- HOMESV (`ubuntu`, Ubuntu 24.04.5 LTS, x86_64) is the **primary dev environment**; repository at `/home/ubuntu/n8n2erpnext/light-luna`. Main was cloned from GitHub and matched SHA and 19 files.
- Development branch: `feat/p0-standalone-foundation`. No production deployments or firewall changes.
- Original release intent says repository should be private; GitHub metadata showed `visibility=public` on 2026-10-13. Owner should review visibility before adding sensitive code or telemetry. No private logs/secrets should be committed.

## Toolchain and constraints

- Go 1.27.2 from official Go download metadata, archive SHA-256 `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`.
- Installed ONLY in user-space `/home/ubuntu/.local/opt/go1.27.2`; no system package install or shell profile modification.
- System APT installation failed because Light Remote execution exposes /var/lib/dpkg and /var/cache/apt as read-only; **do not bypass machine policy or change mount permissions**.
- Run Go with `PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH`, `GOTOOLCHAIN=local` and `GOPROXY=off` for dependency-free P0.
- Go toolchain can build cross-architecture CLI binaries into `/home/ubuntu/.cache/light-luna/build`; binaries are NOT installable packages and do not enforce anything.

## Source files introduced

- `go.mod`, `internal/policy/{model,validate,preview}.go` — bounded typed JSON policy reader, semantic guardrails, inert preview.
- `cmd/lunactl/main.go` — initial `version`, `firewall policies validate`, `firewall policies preview` (NO apply).
- `schemas/policy-v1alpha1.schema.json` — P0 subset, reject unknown policy properties.
- `testdata/host-baseline.json` — synthetic example with TEST-NET IPs; DO NOT install or apply.
- `internal/policy/policy_test.go` — invalid policy, auto-ban refusal, rescue declaration, preview determinism tests.
- `docs/15-P0-IMPLEMENTATION-NOTES.md` — implemented vs deferred contract; README status updated.

## Verified commands

```sh
cd /home/ubuntu/n8n2erpnext/light-luna
export PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH
export GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0
go fmt ./...
go test -count=1 ./...
go vet ./...
go run ./cmd/lunactl firewall policies validate testdata/host-baseline.json
go run ./cmd/lunactl firewall policies preview testdata/host-baseline.json
GOOS=linux GOARCH=amd64 go build -trimpath -o /home/ubuntu/.cache/light-luna/build/lunactl-amd64 ./cmd/lunactl
GOOS=linux GOARCH=arm64 go build -trimpath -o /home/ubuntu/.cache/light-luna/build/lunactl-arm64 ./cmd/lunactl
```

Previous execution of the above completed with exit code 0. Unit tests passed; this is **not** a kernel/netns test, security certification or G-STANDALONE acceptance.

## High priority next tasks

1. Preserve feature-branch commit and remote sync; **do not merge unreviewed P0 into main**.
2. Add duplicate JSON key rejection and schema conformance checks; fuzz malformed policies and alias limits.
3. Design inert nft compiler output and compiler-specific fixture tests; add isolated network namespaces ONLY after validating permissions and that rules target disposable namespaces, never the parent host.
4. Build independent policy gate/executor IPC and rollback watchdog contracts with intentionally disabled apply.
5. Build standalone HTTP/Unix socket read-only API and CLI status, then safely integrated WebUI.
6. Add CI and release gating; actual packages are not ready.
7. Keep `docs/05-NETBIRD.md` as future research only. No NetBird integration/probing/dependency before explicit owner acceptance of **G-STANDALONE**.

## Explicit operational constraints

Do not modify existing Linux host nft/iptables rules, NetBird, Docker, UFW, Light Remote, Sentinel jobs, Cloud, or production services. Do not run real ban, WAF enforcement, forced updates, agent installs or firewall service start. No raw privileged shell execution delegated to Luna Brain. Status/preview is always non-enforcing.

## SSH / Git publishing

- VPS-ARM has existing deploy key allowed for `git@github.com:n8n2erpnext/light-luna.git` and has already published `main`.
- HOMESV clone remote uses the same SSH URL but HOMESV does not yet have an enrolled dedicated deploy key. **Do not copy VPS-ARM private SSH key**.
- To maintain exact branch SHA while HOMESV SSH auth is pending, use a checksum-verified Git bundle transfer from HOMESV to VPS-ARM through approved Light Remote SCP and publish via VPS-ARM key; HOMESV remains authoring/build/test source.
- Long term, generate a separate HOMESV deploy key and have the owner grant repository write permission, then use HOMESV for push directly. No Git credential is embedded in repo.

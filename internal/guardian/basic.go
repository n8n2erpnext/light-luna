// Package guardian reserves a separate self-integrity trust domain.
// P0 does not inspect files, processes or the kernel; report UNKNOWN.
package guardian

type Health struct {
	Status      string `json:"status"`
	Coverage    string `json:"coverage"`
	FileScanner string `json:"fileScanner"`
	EBPF        string `json:"ebpf"`
	Explanation string `json:"explanation"`
}

func Snapshot() Health {
	return Health{Status: "UNKNOWN", Coverage: "NOT_COLLECTED", FileScanner: "NOT_INSTALLED", EBPF: "NOT_INSTALLED", Explanation: "No host evidence collected; this snapshot cannot attest malware absence."}
}

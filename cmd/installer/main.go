// Package main is the OS installer entrypoint for Vyntrio OS.
// Contract: docs/ADR/0007-appliance-installer-contract.md (Block 10).
// Preflight: scripts/installer-preflight.sh (Slice 10.3, read-only).
// Mutation stub: scripts/installer-mutation-stub.sh (Slice 10.5, dry-run only).
// Directory mutation: scripts/installer-mutate-directories.sh (Slice 10.6, target-sandbox only).
// Payload copy: scripts/installer-copy-payloads.sh (Slice 10.7, target-sandbox only).
// Service prep: scripts/installer-prepare-service.sh (Slice 10.8, no service start).
// Bootstrap handoff is deferred to a future Block 10 slice.
package main

func main() {
	// Intentionally empty — foundation phase only.
}

// Package main is the OS installer entrypoint for Vyntrio OS.
// Contract: docs/ADR/0007-appliance-installer-contract.md (Block 10).
// Preflight: scripts/installer-preflight.sh (Slice 10.3, read-only).
// Mutation stub: scripts/installer-mutation-stub.sh (Slice 10.5, dry-run only).
// Directory mutation: scripts/installer-mutate-directories.sh (Slice 10.6, target-sandbox only).
// Payload copy and host target-disk install are deferred to future slices.
package main

func main() {
	// Intentionally empty — foundation phase only.
}

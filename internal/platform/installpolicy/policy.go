package installpolicy

// PolicyVersion identifies the staged installer policy document generation.
const PolicyVersion = "vyntrio-installer-policy-v1"

// Command describes one operator-facing installer tool or subcommand.
type Command struct {
	Name                 string
	Binary               string
	Mutates              bool
	RequiresForce        bool
	RequiresTargetDiskID bool
	Summary              string
	Preconditions        []string
	NotGoals             []string
}

// Commands returns the supported staged pipeline commands in recommended order.
func Commands() []Command {
	return []Command{
		{
			Name:    "verify-artifact",
			Binary:  "vyntrio-verify-artifact",
			Summary: "Read-only integrity check of release manifest artifacts (SHA-256, size).",
			Preconditions: []string{
				"Local release manifest path",
				"Artifact files on disk (optional --base-dir)",
			},
			NotGoals: []string{
				"Does not write disks or install payloads",
				"Does not create USB/ISO media or replace the USB-first product path",
				"Does not replace installer preflight or install/apply",
				"Ed25519 signature verification not implemented",
			},
		},
		{
			Name:                 "preflight",
			Binary:               "vyntrio-installer",
			Summary:              "Read-only target eligibility and optional media verification (Block 10 infra).",
			RequiresTargetDiskID: true,
			Preconditions: []string{
				"Explicit --target-disk-id (opaque ID from storage inventory)",
				"Optional --envelope-root and/or --release-manifest",
			},
			NotGoals: []string{
				"Does not write files or mount disks",
				"Does not auto-select a target disk",
				"Not the primary USB-first delivery path",
			},
		},
		{
			Name:                 "install",
			Binary:               "vyntrio-installer",
			Mutates:              true,
			RequiresForce:        true,
			RequiresTargetDiskID: true,
			Summary:              "Dev/lab only: copy six allowlisted payloads to sandbox (not primary product path).",
			Preconditions: []string{
				"Successful preflight strongly recommended before first run",
				"--force required",
				"--target-disk-id required",
				"--envelope-root and/or --release-manifest required",
			},
			NotGoals: []string{
				"Sandbox install does not mutate block devices",
				"Not a full OS or appliance deployment",
				"Not the primary USB-first operator journey",
				"Does not enable services or bootstrap",
			},
		},
		{
			Name:                 "apply",
			Binary:               "vyntrio-installer",
			Mutates:              true,
			RequiresForce:        true,
			RequiresTargetDiskID: true,
			Summary:              "Live-session/lab infra: apply six allowlisted payloads to one existing partition.",
			Preconditions: []string{
				"Successful preflight required (re-run inside apply)",
				"--force required",
				"--target-disk-id required",
				"--envelope-root and/or --release-manifest required",
				"Exactly one eligible unmounted partition with ext4/xfs/btrfs",
			},
			NotGoals: []string{
				"Never uses sandbox; never falls back to sandbox",
				"Does not create or format partitions",
				"Not USB/ISO delivery; not a full OS install; six payloads only",
			},
		},
		{
			Name:    "postflight",
			Binary:  "vyntrio-installer",
			Summary: "Replay HANDOVER_RECORD.txt from a prior install or apply run (infra audit).",
			Preconditions: []string{
				"--target-root directory containing HANDOVER_RECORD.txt",
				"Sandbox path or apply record directory from a prior command",
			},
			NotGoals: []string{
				"Does not mutate disks or copy payloads",
				"Does not complete OS deployment or USB-first delivery",
			},
		},
	}
}

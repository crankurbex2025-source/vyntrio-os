// Package main is the cross-platform Vyntrio install-media USB writer and media creator.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/mediacreator"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/writemedia"
)

func main() {
	os.Exit(run(os.Args))
}

func run(argv []string) int {
	args := argv[1:]
	// No-args always launches the GUI. Name-based detection was a Windows failure
	// mode (renamed downloads / short names exited with usage and looked "broken").
	if len(args) == 0 {
		return runGUI(nil)
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage()
		return 0
	case "gui":
		return runGUI(args[1:])
	case "list":
		return runList(args[1:])
	case "info":
		return runInfo(args[1:])
	case "write":
		return runWrite(args[1:])
	case "verify-image":
		return runVerifyImage(args[1:])
	case "verify-device":
		return runVerifyDevice(args[1:])
	default:
		usage()
		return 2
	}
}

func runGUI(args []string) int {
	fs := flag.NewFlagSet("gui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	listen := fs.String("listen", "127.0.0.1:17823", "loopback listen address")
	image := fs.String("image", "", "optional default image path")
	noBrowser := fs.Bool("no-browser", false, "do not open a browser window")
	version := fs.String("version", "0.2.0-dev", "version shown in the GUI")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := mediacreator.Run(ctx, mediacreator.Options{
		Version:     *version,
		Listen:      *listen,
		OpenBrowser: !*noBrowser,
		ImageHint:   *image,
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "media-creator gui: %v\n", err)
		return 1
	}
	return 0
}

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	devices, err := writemedia.ListDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-media list: %v\n", err)
		return 1
	}
	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "write-media list: no removable devices found")
		return 0
	}
	for _, device := range devices {
		mounted := "no"
		if device.Mounted {
			mounted = "yes"
		}
		fmt.Printf("id=%s path=%s name=%q size=%d removable=%t bus=%s mounted=%s",
			device.ID, device.Path, device.Name, device.SizeBytes, device.Removable, device.BusType, mounted)
		if len(device.MountPoints) > 0 {
			fmt.Printf(" mount_points=%q", strings.Join(device.MountPoints, ","))
		}
		fmt.Println()
	}
	return 0
}

func runInfo(args []string) int {
	fs := flag.NewFlagSet("info", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imagePath := fs.String("image", "", "path to vyntrio-install-media.img")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*imagePath) == "" {
		fmt.Fprintln(os.Stderr, "write-media info: --image is required")
		return 2
	}
	img, err := writemedia.LoadImage(*imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-media info: %v\n", err)
		return 1
	}
	fmt.Printf("name=%s\nsize=%d\nsha256=%s\n", img.Name, img.SizeBytes, img.SHA256)
	if img.ExpectedSHA256 != "" {
		fmt.Printf("manifest_sha256=%s\n", img.ExpectedSHA256)
	}
	if img.ManifestPath != "" {
		fmt.Printf("manifest=%s\n", img.ManifestPath)
	}
	return 0
}

func runWrite(args []string) int {
	fs := flag.NewFlagSet("write", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imagePath := fs.String("image", "", "path to vyntrio-install-media.img")
	devicePath := fs.String("device", "", "target device path or id from 'list'")
	yes := fs.Bool("yes", false, "confirm destructive write")
	dryRun := fs.Bool("dry-run", false, "print actions without writing")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*imagePath) == "" || strings.TrimSpace(*devicePath) == "" {
		fmt.Fprintln(os.Stderr, "write-media write: --image and --device are required")
		return 2
	}
	if !*yes && !*dryRun {
		fmt.Fprintf(os.Stderr, "write-media write: destructive — re-run with --yes after checking 'list' output\n")
		return 2
	}
	result, err := writemedia.WriteImage(*imagePath, *devicePath, writemedia.WriteOptions{
		AssumeYes: *yes,
		DryRun:    *dryRun,
	})
	if err != nil {
		if errors.Is(err, writemedia.ErrDestructiveWrite) {
			fmt.Fprintln(os.Stderr, "write-media write: confirmation required (--yes)")
			return 2
		}
		fmt.Fprintf(os.Stderr, "write-media write: %v\n", err)
		return 1
	}
	fmt.Printf("write-media: complete device=%s bytes=%d verified=%t\n", result.DevicePath, result.BytesWritten, result.Verified)
	fmt.Fprintln(os.Stderr, "write-media: boot the target device in UEFI or BIOS/legacy mode (dual-mode)")
	return 0
}

func runVerifyImage(args []string) int {
	fs := flag.NewFlagSet("verify-image", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imagePath := fs.String("image", "", "path to vyntrio-install-media.img")
	expected := fs.String("expected-sha256", "", "optional expected digest")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*imagePath) == "" {
		fmt.Fprintln(os.Stderr, "write-media verify-image: --image is required")
		return 2
	}
	if err := writemedia.VerifyImageFile(*imagePath, *expected); err != nil {
		fmt.Fprintf(os.Stderr, "write-media verify-image: %v\n", err)
		return 1
	}
	return 0
}

func runVerifyDevice(args []string) int {
	fs := flag.NewFlagSet("verify-device", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imagePath := fs.String("image", "", "source image path")
	devicePath := fs.String("device", "", "written device path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*imagePath) == "" || strings.TrimSpace(*devicePath) == "" {
		fmt.Fprintln(os.Stderr, "write-media verify-device: --image and --device are required")
		return 2
	}
	if err := writemedia.VerifyDevice(*imagePath, *devicePath); err != nil {
		fmt.Fprintf(os.Stderr, "write-media verify-device: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, strings.TrimSpace(`
vyntrio-write-media / vyntrio-media-creator — USB writer for the Vyntrio BIOS raw install image

GUI (local web wizard on loopback):
  vyntrio-media-creator
  vyntrio-write-media gui [--image <path>] [--listen 127.0.0.1:0] [--no-browser]

CLI:
  vyntrio-write-media list
  vyntrio-write-media info --image <path>
  vyntrio-write-media write --image <path> --device <id-or-path> [--dry-run] [--yes]
  vyntrio-write-media verify-image --image <path> [--expected-sha256 <hex>]
  vyntrio-write-media verify-device --image <path> --device <id-or-path>

This ships a local web GUI (not Electron/Tauri/Qt).
Windows GUI uses subsystem WINDOWS + MessageBox. macOS helpers are Terminal binaries.
Linux is a .tar.gz. Unsigned. Requires elevation to write. BIOS/legacy boot image only.`))
}

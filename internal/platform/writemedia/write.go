package writemedia

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ProgressFunc reports write progress. bytesDone and totalBytes are image sizes.
type ProgressFunc func(bytesDone, totalBytes uint64)

// WriteOptions controls a destructive device write.
type WriteOptions struct {
	AssumeYes bool
	DryRun    bool
	OnProgress ProgressFunc
}

// WriteImage copies the BIOS raw image to the selected device.
func WriteImage(imagePath, devicePath string, opts WriteOptions) (WriteResult, error) {
	devicePath = strings.TrimSpace(devicePath)
	if devicePath == "" {
		return WriteResult{}, fmt.Errorf("device path is required")
	}

	img, err := LoadImage(imagePath)
	if err != nil {
		return WriteResult{}, err
	}

	devices, err := ListDevices()
	if err != nil {
		return WriteResult{}, err
	}
	var selected *Device
	for i := range devices {
		if devices[i].Path == devicePath || devices[i].ID == devicePath {
			selected = &devices[i]
			break
		}
	}
	if selected == nil {
		return WriteResult{}, fmt.Errorf("device %q not found in list output; run 'list' first", devicePath)
	}
	if selected.Mounted {
		return WriteResult{}, fmt.Errorf("device %q appears mounted at %v; unmount first", selected.Path, selected.MountPoints)
	}
	if selected.SizeBytes > 0 && selected.SizeBytes < img.SizeBytes {
		return WriteResult{}, fmt.Errorf("device %q is too small (%d bytes) for image (%d bytes)", selected.Path, selected.SizeBytes, img.SizeBytes)
	}

	fmt.Fprintf(os.Stderr, "write-media: image=%s (%d bytes, SHA-256=%s)\n", img.Name, img.SizeBytes, img.SHA256)
	fmt.Fprintf(os.Stderr, "write-media: target=%s (%s, %d bytes)\n", selected.Path, selected.Name, selected.SizeBytes)
	fmt.Fprintf(os.Stderr, "write-media: WARNING — ALL DATA ON %s WILL BE ERASED\n", selected.Path)

	if opts.DryRun {
		fmt.Fprintf(os.Stderr, "write-media: dry-run — no data written\n")
		return WriteResult{DevicePath: selected.Path, ImagePath: img.Path}, nil
	}
	if !opts.AssumeYes {
		return WriteResult{}, ErrDestructiveWrite
	}

	written, err := writeToDevice(img.Path, selected.Path, img.SizeBytes, opts.OnProgress)
	if err != nil {
		return WriteResult{}, err
	}
	if err := syncDevice(selected.Path); err != nil {
		return WriteResult{}, err
	}

	verified := false
	if err := VerifyDevice(img.Path, selected.Path); err == nil {
		verified = true
	} else {
		fmt.Fprintf(os.Stderr, "write-media: post-write verify failed: %v\n", err)
	}

	return WriteResult{
		DevicePath:   selected.Path,
		ImagePath:    img.Path,
		BytesWritten: written,
		Verified:     verified,
	}, nil
}

func writeToDevice(imagePath, devicePath string, size uint64, onProgress ProgressFunc) (uint64, error) {
	src, err := os.Open(imagePath)
	if err != nil {
		return 0, fmt.Errorf("open image: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := openDeviceForWrite(devicePath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = dst.Close() }()

	reader := io.LimitReader(src, int64(size))
	if onProgress == nil {
		written, err := io.Copy(dst, reader)
		if err != nil {
			return 0, fmt.Errorf("write device: %w", err)
		}
		return uint64(written), nil
	}

	const chunkSize = 4 * 1024 * 1024
	buf := make([]byte, chunkSize)
	var written uint64
	onProgress(0, size)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			wn, writeErr := dst.Write(buf[:n])
			written += uint64(wn)
			if writeErr != nil {
				return written, fmt.Errorf("write device: %w", writeErr)
			}
			if wn != n {
				return written, fmt.Errorf("write device: short write")
			}
			onProgress(written, size)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return written, fmt.Errorf("read image: %w", readErr)
		}
	}
	return written, nil
}

use serde::Serialize;
use sha2::{Digest, Sha256};
use std::fs::{self, File};
use std::io::{Read, Write};
use std::path::Path;
use tauri::{AppHandle, Emitter};

#[derive(Debug, Clone, Serialize)]
pub struct WriteProgress {
    pub bytes_written: u64,
    pub total_bytes: u64,
    pub percent: f64,
    pub phase: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct WriteResult {
    pub device_path: String,
    pub bytes_written: u64,
    pub verified: bool,
    pub image_sha256: String,
    pub message: String,
}

#[derive(Debug)]
pub struct WriteError(pub String);

impl WriteError {
    pub fn msg(s: impl Into<String>) -> Self {
        Self(s.into())
    }
}

impl serde::Serialize for WriteError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.0)
    }
}

pub fn sha256_file(path: &Path) -> Result<String, WriteError> {
    let mut file = File::open(path).map_err(|e| WriteError::msg(format!("open image: {e}")))?;
    let mut hasher = Sha256::new();
    let mut buf = [0u8; 1024 * 1024];
    loop {
        let n = file
            .read(&mut buf)
            .map_err(|e| WriteError::msg(format!("read image: {e}")))?;
        if n == 0 {
            break;
        }
        hasher.update(&buf[..n]);
    }
    Ok(format!("{:x}", hasher.finalize()))
}

pub fn write_image_to_device(
    app: &AppHandle,
    image_path: &str,
    device_path: &str,
    expected_sha256: Option<&str>,
) -> Result<WriteResult, WriteError> {
    let image = Path::new(image_path);
    if !image.is_file() {
        return Err(WriteError::msg(format!("image not found: {image_path}")));
    }
    if device_path.trim().is_empty() {
        return Err(WriteError::msg("device path is required"));
    }
    #[cfg(target_os = "linux")]
    {
        if !device_path.starts_with("/dev/") {
            return Err(WriteError::msg("linux device path must start with /dev/"));
        }
        // Refuse obvious system disks.
        if device_path == "/dev/sda"
            || device_path == "/dev/nvme0n1"
            || device_path.starts_with("/dev/nvme0n1")
        {
            // Still allow if explicitly removable listing selected it — only hard-refuse root fs device.
        }
        if is_root_disk(device_path) {
            return Err(WriteError::msg(
                "refusing to write to the system/root disk — choose a removable USB device",
            ));
        }
    }

    emit(
        app,
        WriteProgress {
            bytes_written: 0,
            total_bytes: 0,
            percent: 0.0,
            phase: "hashing".into(),
        },
    );

    let digest = sha256_file(image)?;
    if let Some(expected) = expected_sha256 {
        let expected = expected.trim().to_lowercase();
        if !expected.is_empty() && expected != digest {
            return Err(WriteError::msg(format!(
                "SHA-256 mismatch: expected {expected}, got {digest}"
            )));
        }
    }

    let total = fs::metadata(image)
        .map_err(|e| WriteError::msg(e.to_string()))?
        .len();

    emit(
        app,
        WriteProgress {
            bytes_written: 0,
            total_bytes: total,
            percent: 0.0,
            phase: "writing".into(),
        },
    );

    let mut src = File::open(image).map_err(|e| WriteError::msg(format!("open image: {e}")))?;
    let mut dst = open_device(device_path)?;

    let mut buf = vec![0u8; 4 * 1024 * 1024];
    let mut written: u64 = 0;
    loop {
        let n = src
            .read(&mut buf)
            .map_err(|e| WriteError::msg(format!("read: {e}")))?;
        if n == 0 {
            break;
        }
        dst.write_all(&buf[..n])
            .map_err(|e| WriteError::msg(format!("write device: {e}")))?;
        written += n as u64;
        let percent = if total > 0 {
            (written as f64 / total as f64) * 100.0
        } else {
            0.0
        };
        emit(
            app,
            WriteProgress {
                bytes_written: written,
                total_bytes: total,
                percent,
                phase: "writing".into(),
            },
        );
    }
    dst.flush()
        .map_err(|e| WriteError::msg(format!("flush: {e}")))?;
    sync_device(device_path)?;

    emit(
        app,
        WriteProgress {
            bytes_written: written,
            total_bytes: total,
            percent: 100.0,
            phase: "verifying".into(),
        },
    );

    let verified = verify_prefix(device_path, image, written.min(total))?;

    emit(
        app,
        WriteProgress {
            bytes_written: written,
            total_bytes: total,
            percent: 100.0,
            phase: "done".into(),
        },
    );

    Ok(WriteResult {
        device_path: device_path.to_string(),
        bytes_written: written,
        verified,
        image_sha256: digest,
        message: if verified {
            "Write completed and prefix verification passed. Boot the USB in UEFI or BIOS/legacy mode (dual-mode image)."
                .into()
        } else {
            "Write completed but prefix verification failed — re-check the device and try again."
                .into()
        },
    })
}

fn emit(app: &AppHandle, progress: WriteProgress) {
    let _ = app.emit("write-progress", progress);
}

fn open_device(device_path: &str) -> Result<File, WriteError> {
    fs::OpenOptions::new()
        .write(true)
        .open(device_path)
        .map_err(|e| {
            WriteError::msg(format!(
                "open device {device_path}: {e} (run the Media Creator elevated / as Administrator)"
            ))
        })
}

fn sync_device(device_path: &str) -> Result<(), WriteError> {
    #[cfg(target_os = "linux")]
    {
        let _ = std::process::Command::new("sync").status();
        let _ = device_path;
        Ok(())
    }
    #[cfg(not(target_os = "linux"))]
    {
        let _ = device_path;
        Ok(())
    }
}

fn verify_prefix(device_path: &str, image: &Path, bytes: u64) -> Result<bool, WriteError> {
    let check_len = bytes.min(1024 * 1024);
    let mut img = File::open(image).map_err(|e| WriteError::msg(e.to_string()))?;
    let mut dev = File::open(device_path).map_err(|e| WriteError::msg(format!("reopen device: {e}")))?;
    let mut a = vec![0u8; check_len as usize];
    let mut b = vec![0u8; check_len as usize];
    img.read_exact(&mut a)
        .map_err(|e| WriteError::msg(format!("verify read image: {e}")))?;
    dev.read_exact(&mut b)
        .map_err(|e| WriteError::msg(format!("verify read device: {e}")))?;
    Ok(a == b)
}

#[cfg(target_os = "linux")]
fn is_root_disk(device_path: &str) -> bool {
    let Ok(mounts) = fs::read_to_string("/proc/mounts") else {
        return false;
    };
    for line in mounts.lines() {
        let mut parts = line.split_whitespace();
        let Some(src) = parts.next() else { continue };
        let Some(dst) = parts.next() else { continue };
        if dst == "/" {
            // /dev/sda2 -> /dev/sda
            if let Some(base) = src.strip_prefix("/dev/") {
                let disk = base
                    .trim_end_matches(|c: char| c.is_ascii_digit())
                    .trim_end_matches(|c: char| c == 'p' && base.contains("nvme"));
                let candidate = format!("/dev/{disk}");
                // Better: compare prefixes carefully.
                if device_path == src || src.starts_with(device_path) || device_path.starts_with(&format!("{src}")) {
                    return true;
                }
                // nvme0n1p2 -> nvme0n1
                if device_path == candidate || src.starts_with(device_path) {
                    return true;
                }
            }
            if src.starts_with(device_path) {
                return true;
            }
        }
    }
    false
}

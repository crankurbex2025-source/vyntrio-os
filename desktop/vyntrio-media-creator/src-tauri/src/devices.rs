use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

#[cfg(any(target_os = "macos", target_os = "windows"))]
use std::process::Command;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StorageDevice {
    pub id: String,
    pub path: String,
    pub name: String,
    pub size_bytes: u64,
    pub removable: bool,
    pub bus_type: String,
    pub mounted: bool,
    pub mount_points: Vec<String>,
}

#[derive(Debug)]
pub struct DeviceError(pub String);

impl DeviceError {
    pub fn msg(s: impl Into<String>) -> Self {
        Self(s.into())
    }
}

impl serde::Serialize for DeviceError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.0)
    }
}

pub fn list_removable_devices() -> Result<Vec<StorageDevice>, DeviceError> {
    #[cfg(target_os = "linux")]
    {
        list_linux()
    }
    #[cfg(target_os = "macos")]
    {
        list_macos()
    }
    #[cfg(target_os = "windows")]
    {
        list_windows()
    }
    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    {
        Err(DeviceError::msg("unsupported platform for device listing"))
    }
}

#[cfg(target_os = "linux")]
fn list_linux() -> Result<Vec<StorageDevice>, DeviceError> {
    let entries = fs::read_dir("/sys/block").map_err(|e| DeviceError::msg(e.to_string()))?;
    let mut devices = Vec::new();
    for entry in entries.flatten() {
        let name = entry.file_name().to_string_lossy().to_string();
        if name.starts_with("loop")
            || name.starts_with("ram")
            || name.starts_with("dm-")
            || name.starts_with("zram")
            || name.starts_with("md")
        {
            continue;
        }
        let removable = read_sys(&name, "removable") == "1";
        let usb = is_usb(&name);
        if !removable && !usb {
            continue;
        }
        let size_sectors: u64 = read_sys(&name, "size").parse().unwrap_or(0);
        let mut model = read_sys(&name, "device/model");
        if model.is_empty() {
            model = name.clone();
        }
        let path = format!("/dev/{name}");
        let (mounted, mounts) = mount_info(&path);
        devices.push(StorageDevice {
            id: path.clone(),
            path,
            name: model.trim().to_string(),
            size_bytes: size_sectors.saturating_mul(512),
            removable,
            bus_type: if usb { "usb".into() } else { "block".into() },
            mounted,
            mount_points: mounts,
        });
    }
    Ok(devices)
}

#[cfg(target_os = "linux")]
fn read_sys(block: &str, rel: &str) -> String {
    fs::read_to_string(format!("/sys/block/{block}/{rel}"))
        .unwrap_or_default()
        .trim()
        .to_string()
}

#[cfg(target_os = "linux")]
fn is_usb(block: &str) -> bool {
    fs::read_link(format!("/sys/block/{block}/device"))
        .map(|p| p.to_string_lossy().contains("usb"))
        .unwrap_or(false)
}

#[cfg(target_os = "linux")]
fn mount_info(device_path: &str) -> (bool, Vec<String>) {
    let Ok(mounts) = fs::read_to_string("/proc/mounts") else {
        return (false, vec![]);
    };
    let mut points = Vec::new();
    for line in mounts.lines() {
        let mut parts = line.split_whitespace();
        let Some(src) = parts.next() else { continue };
        let Some(dst) = parts.next() else { continue };
        if src == device_path || src.starts_with(&format!("{device_path}")) {
            points.push(dst.to_string());
        }
    }
    (!points.is_empty(), points)
}

#[cfg(target_os = "macos")]
fn list_macos() -> Result<Vec<StorageDevice>, DeviceError> {
    let output = Command::new("diskutil")
        .args(["list", "-plist", "external", "physical"])
        .output()
        .map_err(|e| DeviceError::msg(format!("diskutil: {e}")))?;
    if !output.status.success() {
        // Fall back to full list and filter.
        return list_macos_fallback();
    }
    // Prefer `diskutil info -all` parsing via simple `diskutil list external`.
    list_macos_fallback()
}

#[cfg(target_os = "macos")]
fn list_macos_fallback() -> Result<Vec<StorageDevice>, DeviceError> {
    let output = Command::new("diskutil")
        .args(["list", "external", "physical"])
        .output()
        .map_err(|e| DeviceError::msg(format!("diskutil: {e}")))?;
    let text = String::from_utf8_lossy(&output.stdout);
    let mut devices = Vec::new();
    for line in text.lines() {
        let line = line.trim();
        if let Some(rest) = line.strip_prefix("/dev/disk") {
            let id: String = rest
                .chars()
                .take_while(|c| c.is_ascii_digit())
                .collect();
            if id.is_empty() {
                continue;
            }
            let path = format!("/dev/disk{id}");
            let info = Command::new("diskutil")
                .args(["info", &path])
                .output()
                .ok();
            let info_text = info
                .as_ref()
                .map(|o| String::from_utf8_lossy(&o.stdout).to_string())
                .unwrap_or_default();
            let name = info_field(&info_text, "Device / Media Name").unwrap_or_else(|| path.clone());
            let size = parse_macos_size(&info_text).unwrap_or(0);
            let mounted = info_text.contains("Mounted:               Yes");
            devices.push(StorageDevice {
                id: path.clone(),
                path,
                name,
                size_bytes: size,
                removable: true,
                bus_type: "usb".into(),
                mounted,
                mount_points: vec![],
            });
        }
    }
    Ok(devices)
}

#[cfg(target_os = "macos")]
fn info_field(text: &str, key: &str) -> Option<String> {
    for line in text.lines() {
        let line = line.trim();
        if let Some(rest) = line.strip_prefix(key) {
            let rest = rest.trim().trim_start_matches(':').trim();
            if !rest.is_empty() {
                return Some(rest.to_string());
            }
        }
    }
    None
}

#[cfg(target_os = "macos")]
fn parse_macos_size(text: &str) -> Option<u64> {
    let disk_size = info_field(text, "Disk Size")?;
    // e.g. "15.5 GB (15518924800 Bytes) (exactly ...)"
    if let Some(start) = disk_size.find('(') {
        let rest = &disk_size[start + 1..];
        let num: String = rest.chars().take_while(|c| c.is_ascii_digit()).collect();
        return num.parse().ok();
    }
    None
}

#[cfg(target_os = "windows")]
fn list_windows() -> Result<Vec<StorageDevice>, DeviceError> {
    let script = r#"
Get-CimInstance Win32_DiskDrive | Where-Object { $_.InterfaceType -eq 'USB' -or $_.MediaType -match 'Removable' } | ForEach-Object {
  $idx = $_.Index
  $size = [uint64]$_.Size
  $model = $_.Model
  $path = "\\.\PHYSICALDRIVE$idx"
  [PSCustomObject]@{ id=$path; path=$path; name=$model; size_bytes=$size; removable=$true; bus_type='usb'; mounted=$false; mount_points=@() } | ConvertTo-Json -Compress
}
"#;
    let output = Command::new("powershell")
        .args(["-NoProfile", "-Command", script])
        .output()
        .map_err(|e| DeviceError::msg(format!("powershell: {e}")))?;
    if !output.status.success() {
        return Err(DeviceError::msg(format!(
            "powershell failed: {}",
            String::from_utf8_lossy(&output.stderr)
        )));
    }
    let text = String::from_utf8_lossy(&output.stdout);
    let mut devices = Vec::new();
    for line in text.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }
        if let Ok(dev) = serde_json::from_str::<StorageDevice>(line) {
            devices.push(dev);
        }
    }
    Ok(devices)
}

pub fn suggest_image_paths() -> Vec<String> {
    let mut out = Vec::new();
    let mut roots: Vec<PathBuf> = Vec::new();
    if let Ok(cwd) = std::env::current_dir() {
        roots.push(cwd);
    }
    if let Ok(home) = std::env::var("HOME") {
        roots.push(PathBuf::from(&home).join("Downloads"));
        roots.push(PathBuf::from(&home).join("Desktop"));
    }
    if let Ok(userprofile) = std::env::var("USERPROFILE") {
        roots.push(PathBuf::from(&userprofile).join("Downloads"));
        roots.push(PathBuf::from(&userprofile).join("Desktop"));
    }
    // Repo-relative staging when running from a checkout.
    roots.push(PathBuf::from("/opt/vyntrio-os/distro/release/staging"));
    for root in roots {
        if let Ok(rd) = fs::read_dir(&root) {
            for entry in rd.flatten() {
                let path = entry.path();
                let name = path.file_name().and_then(|s| s.to_str()).unwrap_or("");
                if name.starts_with("vyntrio-install-media") && name.ends_with(".img") {
                    out.push(path.to_string_lossy().to_string());
                }
            }
        }
    }
    out.sort();
    out.dedup();
    out
}

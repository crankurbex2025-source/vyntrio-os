use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReleaseArtifact {
    pub name: String,
    pub format: String,
    pub firmware_boot_mode: String,
    pub bios_support: bool,
    pub uefi_support: bool,
    pub dual_mode: bool,
    pub secure_boot: Option<String>,
    pub media_role: Option<String>,
    pub size_bytes: Option<u64>,
    pub sha256: Option<String>,
    pub download_available: bool,
    pub download_path: Option<String>,
    pub version: String,
    pub build_id: Option<String>,
    pub channel: Option<String>,
    pub support_status: Option<String>,
    pub generated_at: Option<String>,
    pub source: String,
    pub latest: bool,
}

#[derive(Debug, Deserialize)]
struct PublicInstallMedia {
    generated_at: Option<String>,
    release: ReleaseLine,
    primary_artifact: PrimaryArtifact,
    #[serde(default)]
    image_versions: Vec<ImageVersionEntry>,
    support_status: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ReleaseLine {
    version: String,
    channel: Option<String>,
    build_id: Option<String>,
}

#[derive(Debug, Deserialize)]
struct PrimaryArtifact {
    name: String,
    format: String,
    firmware_boot_mode: String,
    bios_support: Option<bool>,
    uefi_support: Option<bool>,
    dual_mode: Option<bool>,
    secure_boot: Option<String>,
    media_role: Option<String>,
    size_bytes: Option<u64>,
    sha256: Option<String>,
    download_available: Option<bool>,
    download_path: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ImageVersionEntry {
    version: String,
    build_id: Option<String>,
    channel: Option<String>,
    generated_at: Option<String>,
    name: String,
    format: String,
    size_bytes: Option<u64>,
    sha256: Option<String>,
    firmware_boot_mode: String,
    bios_support: Option<bool>,
    uefi_support: Option<bool>,
    dual_mode: Option<bool>,
    secure_boot: Option<String>,
    download_available: Option<bool>,
    download_path: Option<String>,
    support_status: Option<String>,
    latest: Option<bool>,
    media_role: Option<String>,
}

#[derive(Debug)]
pub struct ReleaseError(pub String);

impl serde::Serialize for ReleaseError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.0)
    }
}

pub fn list_releases(metadata_url: Option<String>) -> Result<Vec<ReleaseArtifact>, ReleaseError> {
    let mut releases = Vec::new();

    if let Some(mut local) = load_local_staging() {
        releases.append(&mut local);
    }

    let url = metadata_url.unwrap_or_else(|| {
        std::env::var("VYNTRIO_INSTALL_MEDIA_API")
            .unwrap_or_else(|_| "http://127.0.0.1:8080/api/v1/public/install-media".into())
    });

    if let Ok(mut remote) = fetch_remote(&url) {
        for r in remote.drain(..) {
            let dup = releases.iter().any(|x| {
                x.sha256.is_some() && x.sha256 == r.sha256 && r.sha256.is_some()
            });
            if !dup {
                releases.push(r);
            }
        }
    }

    if releases.is_empty() {
        releases.push(fallback_release());
    }

    // Prefer latest first, then by generated_at/version.
    releases.sort_by(|a, b| b.latest.cmp(&a.latest).then(b.version.cmp(&a.version)));

    // Deduplicate by sha/version.
    let mut seen = std::collections::HashSet::new();
    releases.retain(|r| {
        let key = format!(
            "{}:{}",
            r.sha256.clone().unwrap_or_default(),
            r.version
        );
        seen.insert(key)
    });

    Ok(releases)
}

fn fallback_release() -> ReleaseArtifact {
    ReleaseArtifact {
        name: "vyntrio-install-media.img".into(),
        format: "raw_gpt_hybrid_appliance".into(),
        firmware_boot_mode: "bios+uefi".into(),
        bios_support: true,
        uefi_support: true,
        dual_mode: true,
        secure_boot: Some("unsupported".into()),
        media_role: Some("appliance".into()),
        size_bytes: None,
        sha256: None,
        download_available: false,
        download_path: None,
        version: "0.2.0-dev".into(),
        build_id: None,
        channel: Some("development".into()),
        support_status: Some("engineering_media_early_access".into()),
        generated_at: None,
        source: "bundled_fallback".into(),
        latest: true,
    }
}

fn load_local_staging() -> Option<Vec<ReleaseArtifact>> {
    let path = "/opt/vyntrio-os/distro/release/staging/install-media-public.json";
    let data = std::fs::read_to_string(path).ok()?;
    let parsed: PublicInstallMedia = serde_json::from_str(&data).ok()?;
    Some(expand_versions(parsed, "local_staging"))
}

fn fetch_remote(url: &str) -> Result<Vec<ReleaseArtifact>, ReleaseError> {
    let response = ureq::get(url)
        .call()
        .map_err(|e| ReleaseError(format!("fetch {url}: {e}")))?;
    let parsed: PublicInstallMedia = response
        .into_json()
        .map_err(|e| ReleaseError(format!("decode metadata: {e}")))?;
    Ok(expand_versions(parsed, "remote_api"))
}

fn expand_versions(parsed: PublicInstallMedia, source: &str) -> Vec<ReleaseArtifact> {
    if !parsed.image_versions.is_empty() {
        return parsed
            .image_versions
            .into_iter()
            .map(|v| from_version_entry(v, source))
            .collect();
    }
    vec![to_artifact(parsed, source)]
}

fn from_version_entry(v: ImageVersionEntry, source: &str) -> ReleaseArtifact {
    let mode = v.firmware_boot_mode.clone();
    let bios = v.bios_support.unwrap_or(mode.contains("bios"));
    let uefi = v.uefi_support.unwrap_or(mode.contains("uefi"));
    let dual = v.dual_mode.unwrap_or(mode == "bios+uefi");
    ReleaseArtifact {
        name: v.name,
        format: v.format,
        firmware_boot_mode: mode,
        bios_support: bios,
        uefi_support: uefi,
        dual_mode: dual,
        secure_boot: v.secure_boot.or_else(|| Some("unsupported".into())),
        media_role: v.media_role.or_else(|| Some("appliance".into())),
        size_bytes: v.size_bytes,
        sha256: v.sha256,
        download_available: v.download_available.unwrap_or(false),
        download_path: v.download_path,
        version: v.version,
        build_id: v.build_id,
        channel: v.channel,
        support_status: v.support_status,
        generated_at: v.generated_at,
        source: source.into(),
        latest: v.latest.unwrap_or(false),
    }
}

fn to_artifact(parsed: PublicInstallMedia, source: &str) -> ReleaseArtifact {
    let a = parsed.primary_artifact;
    let mode = a.firmware_boot_mode.clone();
    let bios = a.bios_support.unwrap_or(mode.contains("bios"));
    let uefi = a.uefi_support.unwrap_or(mode.contains("uefi"));
    let dual = a.dual_mode.unwrap_or(mode == "bios+uefi");
    ReleaseArtifact {
        name: a.name,
        format: a.format,
        firmware_boot_mode: mode,
        bios_support: bios,
        uefi_support: uefi,
        dual_mode: dual,
        secure_boot: a.secure_boot.or_else(|| Some("unsupported".into())),
        media_role: a.media_role.or_else(|| Some("appliance".into())),
        size_bytes: a.size_bytes,
        sha256: a.sha256,
        download_available: a.download_available.unwrap_or(false),
        download_path: a.download_path,
        version: parsed.release.version,
        build_id: parsed.release.build_id,
        channel: parsed.release.channel,
        support_status: parsed.support_status,
        generated_at: parsed.generated_at,
        source: source.into(),
        latest: true,
    }
}

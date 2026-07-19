mod devices;
mod releases;
mod write;

use devices::{list_removable_devices, suggest_image_paths, StorageDevice};
use releases::{list_releases, ReleaseArtifact};
use write::{write_image_to_device, WriteResult};

#[tauri::command]
fn get_app_info() -> serde_json::Value {
    serde_json::json!({
        "name": "Vyntrio Media Creator",
        "version": env!("CARGO_PKG_VERSION"),
        "framework": "tauri",
        "kind": "native_desktop",
        "uefi_baseline": "dual_mode_required",
        "platforms": ["windows", "macos", "linux"],
    })
}

#[tauri::command]
fn list_storage_devices() -> Result<Vec<StorageDevice>, String> {
    list_removable_devices().map_err(|e| e.0)
}

#[tauri::command]
fn suggest_images() -> Vec<String> {
    suggest_image_paths()
}

#[tauri::command]
fn list_install_releases(metadata_url: Option<String>) -> Result<Vec<ReleaseArtifact>, String> {
    list_releases(metadata_url).map_err(|e| e.0)
}

#[tauri::command]
fn write_install_media(
    app: tauri::AppHandle,
    image_path: String,
    device_path: String,
    expected_sha256: Option<String>,
) -> Result<WriteResult, String> {
    write_image_to_device(
        &app,
        &image_path,
        &device_path,
        expected_sha256.as_deref(),
    )
    .map_err(|e| e.0)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![
            get_app_info,
            list_storage_devices,
            suggest_images,
            list_install_releases,
            write_install_media
        ])
        .run(tauri::generate_context!())
        .expect("error while running Vyntrio Media Creator");
}

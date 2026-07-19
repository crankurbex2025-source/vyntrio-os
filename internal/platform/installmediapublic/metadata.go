package installmediapublic

const (
	PublicationNotBuilt     = "not_built"
	PublicationLocalStaging = "local_staging"
	PublicationUnavailable  = "unavailable"
	ArtifactName            = "vyntrio-install-media.img"
	ArtifactNameLegacyBIOS  = "vyntrio-install-media-bios.img"
	ManifestName            = "release-manifest.json"
	PublicMetadataName      = "install-media-public.json"
	StagingRecordName       = "STAGING.txt"

	SupportStatusEngineering = "engineering_media_early_access"
)

// Metadata is the public install-media DTO for GET /api/v1/public/install-media.
type Metadata struct {
	PublicationStatus string          `json:"publication_status"`
	GeneratedAt       string          `json:"generated_at,omitempty"`
	Release           ReleaseLine     `json:"release"`
	PrimaryArtifact   PrimaryArtifact `json:"primary_artifact"`
	ImageVersions     []ImageVersion  `json:"image_versions,omitempty"`
	BuildTarget       string          `json:"build_target"`
	StageTarget       string          `json:"stage_target"`
	VerifyCommand     string          `json:"verify_command,omitempty"`
	Limitations       []string        `json:"limitations"`
	SupportStatus     string          `json:"support_status,omitempty"`
	Writer            *WriterInfo     `json:"writer,omitempty"`
}

// ReleaseLine describes the release line for the install image.
type ReleaseLine struct {
	Version string `json:"version"`
	Channel string `json:"channel,omitempty"`
	BuildID string `json:"build_id,omitempty"`
}

// ImageVersion is one published appliance image entry for the Media Creator feed.
type ImageVersion struct {
	Version           string  `json:"version"`
	BuildID           string  `json:"build_id,omitempty"`
	Channel           string  `json:"channel,omitempty"`
	GeneratedAt       string  `json:"generated_at,omitempty"`
	Name              string  `json:"name"`
	Format            string  `json:"format"`
	SizeBytes         uint64  `json:"size_bytes"`
	SHA256            string  `json:"sha256"`
	FirmwareBootMode  string  `json:"firmware_boot_mode"`
	BiosSupport       bool    `json:"bios_support"`
	UefiSupport       bool    `json:"uefi_support"`
	DualMode          bool    `json:"dual_mode"`
	SecureBoot        string  `json:"secure_boot,omitempty"`
	DownloadAvailable bool    `json:"download_available"`
	DownloadPath      string  `json:"download_path,omitempty"`
	SupportStatus     string  `json:"support_status,omitempty"`
	Latest            bool    `json:"latest"`
	MediaRole         string  `json:"media_role,omitempty"`
}

// PrimaryArtifact describes the recommended install image.
type PrimaryArtifact struct {
	Name              string  `json:"name"`
	Format            string  `json:"format"`
	FirmwareBootMode  string  `json:"firmware_boot_mode"`
	BiosSupport       bool    `json:"bios_support"`
	UefiSupport       bool    `json:"uefi_support"`
	DualMode          bool    `json:"dual_mode"`
	SecureBoot        string  `json:"secure_boot,omitempty"`
	MediaRole         string  `json:"media_role,omitempty"`
	SizeBytes         *uint64 `json:"size_bytes,omitempty"`
	SHA256            string  `json:"sha256,omitempty"`
	DownloadAvailable bool    `json:"download_available"`
	DownloadPath      string  `json:"download_path,omitempty"`
	ManifestPath      string  `json:"manifest_path,omitempty"`
}

// WriterInfo describes the cross-platform install-media creator / writer.
type WriterInfo struct {
	Name              string           `json:"name"`
	Kind              string           `json:"kind"`
	Platforms         []string         `json:"platforms"`
	BinaryName        string           `json:"binary_name"`
	BuildTarget       string           `json:"build_target"`
	PackageTarget     string           `json:"package_target,omitempty"`
	DocumentationPath string           `json:"documentation_path"`
	RequiresElevation bool             `json:"requires_elevation"`
	NativeGUI         bool             `json:"native_gui"`
	GUIAvailable      bool             `json:"gui_available"`
	GUIKind           string           `json:"gui_kind,omitempty"`
	Artifacts         []WriterArtifact `json:"artifacts"`
}

// WriterArtifact is one downloadable writer/creator package for a platform.
type WriterArtifact struct {
	Platform          string  `json:"platform"`
	Arch              string  `json:"arch"`
	Name              string  `json:"name"`
	Kind              string  `json:"kind"`
	SizeBytes         *uint64 `json:"size_bytes,omitempty"`
	SHA256            string  `json:"sha256,omitempty"`
	DownloadAvailable bool    `json:"download_available"`
	DownloadPath      string  `json:"download_path,omitempty"`
}

// WriterArtifactNames maps packaged creator/writer filenames to platform/arch.
// Native Tauri packages only for claimed GUI. Withdrawn: Go loopback GUI, macOS helpers-as-GUI, .app.zip.
var WriterArtifactNames = []WriterArtifact{
	{Platform: "windows", Arch: "amd64", Name: "vyntrio-media-creator-windows-amd64-setup.exe", Kind: "native_nsis_installer"},
	{Platform: "linux", Arch: "amd64", Name: "vyntrio-media-creator-linux-amd64.deb", Kind: "native_deb"},
	{Platform: "linux", Arch: "amd64", Name: "vyntrio-media-creator-linux-amd64.AppImage", Kind: "native_appimage"},
	{Platform: "windows", Arch: "amd64", Name: "vyntrio-write-media-windows-amd64.exe", Kind: "cli_binary"},
	{Platform: "macos", Arch: "arm64", Name: "vyntrio-write-media-darwin-arm64", Kind: "cli_binary"},
	{Platform: "macos", Arch: "amd64", Name: "vyntrio-write-media-darwin-amd64", Kind: "cli_binary"},
	{Platform: "linux", Arch: "amd64", Name: "vyntrio-write-media-linux-amd64", Kind: "cli_binary"},
}

// WriterStagingSubdir is the staging subdirectory for writer binaries.
const WriterStagingSubdir = "writer"

func defaultWriterInfo() *WriterInfo {
	artifacts := make([]WriterArtifact, len(WriterArtifactNames))
	copy(artifacts, WriterArtifactNames)
	return &WriterInfo{
		Name:              "vyntrio-media-creator",
		Kind:              "native_desktop_tauri",
		Platforms:         []string{"linux", "windows"},
		BinaryName:        "vyntrio-media-creator",
		BuildTarget:       "make build-media-creator-native",
		PackageTarget:     "make package-media-creator-native",
		DocumentationPath: "docs/ops/install-media-writer.md",
		RequiresElevation: true,
		NativeGUI:         true,
		GUIAvailable:      true,
		GUIKind:           "tauri",
		Artifacts:         artifacts,
	}
}

package system_setting

import "github.com/QuantumNous/new-api/setting/config"

type PlaygroundAttachmentSettings struct {
	Enabled                bool     `json:"enabled"`
	StorageDriver          string   `json:"storage_driver"`
	TTLHours               int      `json:"ttl_hours"`
	MaxFileSizeBytes       int64    `json:"max_file_size_bytes"`
	MaxFilesPerMessage     int      `json:"max_files_per_message"`
	MaxFilesPerSession     int      `json:"max_files_per_session"`
	AllowedMIMETypes       []string `json:"allowed_mime_types"`
	ReferenceTTLSeconds    int      `json:"reference_ttl_seconds"`
	LocalBasePath          string   `json:"local_base_path"`
	CleanupIntervalMinutes int      `json:"cleanup_interval_minutes"`
	CleanupBatchSize       int      `json:"cleanup_batch_size"`
	OSSEndpoint            string   `json:"oss_endpoint"`
	OSSBucket              string   `json:"oss_bucket"`
	OSSRegion              string   `json:"oss_region"`
	OSSAccessKeyID         string   `json:"oss_api_key"`
	OSSAccessKeySecret     string   `json:"oss_secret"`
	OSSObjectPrefix        string   `json:"oss_object_prefix"`
}

var playgroundAttachmentSettings = PlaygroundAttachmentSettings{
	Enabled:                false,
	StorageDriver:          "local",
	TTLHours:               24,
	MaxFileSizeBytes:       10 * 1024 * 1024,
	MaxFilesPerMessage:     4,
	MaxFilesPerSession:     20,
	AllowedMIMETypes:       []string{"image/png", "image/jpeg", "image/gif", "image/webp", "text/plain", "application/pdf"},
	ReferenceTTLSeconds:    300,
	LocalBasePath:          "./data/playground-attachments",
	CleanupIntervalMinutes: 30,
	CleanupBatchSize:       100,
	OSSObjectPrefix:        "playground",
}

func init() {
	config.GlobalConfig.Register("playground_attachment", &playgroundAttachmentSettings)
}

func GetPlaygroundAttachmentSettings() *PlaygroundAttachmentSettings {
	return &playgroundAttachmentSettings
}

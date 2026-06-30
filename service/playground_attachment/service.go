package playground_attachment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/playground_attachment/driver"
	"github.com/QuantumNous/new-api/service/playground_attachment/driver/oss"
	"github.com/QuantumNous/new-api/setting/system_setting"
	_ "golang.org/x/image/webp"
)

const (
	ErrCodeFeatureDisabled          = "attachment_feature_disabled"
	ErrCodeFileTooLarge             = "attachment_file_too_large"
	ErrCodeMIMENotAllowed           = "attachment_mime_not_allowed"
	ErrCodeTooManyAttachments       = "attachment_too_many"
	ErrCodeAttachmentNotFound       = "attachment_not_found"
	ErrCodeAttachmentExpired        = "attachment_expired"
	ErrCodeAccessDenied             = "attachment_access_denied"
	ErrCodeStorageDriverUnavailable = "attachment_storage_driver_unavailable"
	ErrCodeStorageWriteFailed       = "attachment_storage_write_failed"
	ErrCodeStorageReadFailed        = "attachment_storage_read_failed"
	ErrCodeStorageDeleteFailed      = "attachment_storage_delete_failed"
	ErrCodeInvalidSignedReference   = "attachment_invalid_signed_reference"
	ErrCodeInvalidRequest           = "attachment_invalid_request"
)

type Settings struct {
	Enabled             bool
	StorageDriver       string
	TTLHours            int
	MaxFileSizeBytes    int64
	MaxFilesPerMessage  int
	MaxFilesPerSession  int
	AllowedMIMETypes    []string
	ReferenceTTLSeconds int
}

type UploadInput struct {
	UserID      int
	SessionID   string
	Filename    string
	ContentType string
	Reader      io.Reader
}

type ReferenceInput struct {
	UserID        int
	AttachmentIDs []string
	PublicBaseURL string
}

type LocalReferenceInput struct {
	AttachmentID string
	ObjectKey    string
	ExpiresAt    int64
	Signature    string
}

type Reference struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
	Signature string `json:"signature,omitempty"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	Status    string `json:"status"`
}

type CleanupResult struct {
	ExpiredCount int `json:"expired_count"`
	FailedCount  int `json:"failed_count"`
}

type AttachmentError struct {
	Code    string
	Message string
}

func (e *AttachmentError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

type Service struct {
	driver           driver.Driver
	settingsProvider func() Settings
	now              func() time.Time
}

func NewService(storageDriver driver.Driver, settingsProvider func() Settings) *Service {
	return &Service{
		driver:           storageDriver,
		settingsProvider: settingsProvider,
		now:              time.Now,
	}
}

func NewDefaultService() *Service {
	settings := SettingsFromSystem()
	return NewService(newConfiguredDriver(settings), SettingsFromSystem)
}

func SettingsFromSystem() Settings {
	cfg := system_setting.GetPlaygroundAttachmentSettings()
	return Settings{
		Enabled:             cfg.Enabled,
		StorageDriver:       cfg.StorageDriver,
		TTLHours:            cfg.TTLHours,
		MaxFileSizeBytes:    cfg.MaxFileSizeBytes,
		MaxFilesPerMessage:  cfg.MaxFilesPerMessage,
		MaxFilesPerSession:  cfg.MaxFilesPerSession,
		AllowedMIMETypes:    append([]string(nil), cfg.AllowedMIMETypes...),
		ReferenceTTLSeconds: cfg.ReferenceTTLSeconds,
	}
}

func newConfiguredDriver(settings Settings) driver.Driver {
	cfg := system_setting.GetPlaygroundAttachmentSettings()
	switch strings.ToLower(strings.TrimSpace(settings.StorageDriver)) {
	case "", "local":
		return NewLocalDriver(cfg.LocalBasePath)
	case "oss":
		return oss.New(oss.Config{
			Endpoint:        cfg.OSSEndpoint,
			Bucket:          cfg.OSSBucket,
			Region:          cfg.OSSRegion,
			AccessKeyID:     cfg.OSSAccessKeyID,
			AccessKeySecret: cfg.OSSAccessKeySecret,
			ObjectPrefix:    cfg.OSSObjectPrefix,
		})
	default:
		return nil
	}
}

func (s *Service) driverForAttachment(attachment *model.PlaygroundAttachment) driver.Driver {
	if s.driver != nil && s.driver.Name() == attachment.Driver {
		return s.driver
	}
	settings := SettingsFromSystem()
	settings.StorageDriver = attachment.Driver
	return newConfiguredDriver(settings)
}

func (s *Service) resolveAttachmentDriver(attachment *model.PlaygroundAttachment) (driver.Driver, error) {
	storageDriver := s.driverForAttachment(attachment)
	if storageDriver == nil {
		return nil, attachmentError(ErrCodeStorageDriverUnavailable)
	}
	return storageDriver, nil
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (*model.PlaygroundAttachment, error) {
	settings := normalizedSettings(s.settingsProvider())
	if !settings.Enabled {
		return nil, attachmentError(ErrCodeFeatureDisabled)
	}
	if s.driver == nil {
		return nil, attachmentError(ErrCodeStorageDriverUnavailable)
	}
	if input.UserID <= 0 || strings.TrimSpace(input.SessionID) == "" || input.Reader == nil {
		return nil, attachmentError(ErrCodeInvalidRequest)
	}

	now := s.now()
	activeCount, err := model.CountActivePlaygroundAttachmentsForSession(input.UserID, input.SessionID, now.Unix())
	if err != nil {
		return nil, err
	}
	if settings.MaxFilesPerMessage > 0 && activeCount >= int64(settings.MaxFilesPerMessage) {
		return nil, attachmentError(ErrCodeTooManyAttachments)
	}
	if settings.MaxFilesPerSession > 0 && activeCount >= int64(settings.MaxFilesPerSession) {
		return nil, attachmentError(ErrCodeTooManyAttachments)
	}

	data, err := readUploadBytes(input.Reader, settings.MaxFileSizeBytes)
	if err != nil {
		return nil, err
	}
	mimeType := detectUploadMIME(data, input.Filename, input.ContentType)
	if !isMIMEAllowed(mimeType, settings.AllowedMIMETypes) {
		return nil, attachmentError(ErrCodeMIMENotAllowed)
	}
	if strings.HasPrefix(mimeType, "image/") && !isDecodableImage(mimeType, data) {
		return nil, attachmentError(ErrCodeMIMENotAllowed)
	}

	attachmentID, err := newAttachmentID()
	if err != nil {
		return nil, err
	}
	objectKey := newObjectKey(input.UserID, attachmentID, now)
	contentDigest := sha256.Sum256(data)
	attachment := &model.PlaygroundAttachment{
		ID:        attachmentID,
		UserID:    input.UserID,
		SessionID: strings.TrimSpace(input.SessionID),
		Driver:    s.driver.Name(),
		ObjectKey: objectKey,
		Filename:  sanitizeFilename(input.Filename),
		MimeType:  mimeType,
		Size:      int64(len(data)),
		Digest:    hex.EncodeToString(contentDigest[:]),
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(time.Duration(settings.TTLHours) * time.Hour).Unix(),
	}

	written, err := s.driver.Put(ctx, objectKey, mimeType, bytes.NewReader(data))
	if err != nil {
		return nil, attachmentError(ErrCodeStorageWriteFailed)
	}
	if written != attachment.Size {
		_ = s.driver.Delete(ctx, objectKey)
		return nil, attachmentError(ErrCodeStorageWriteFailed)
	}
	if err := model.CreatePlaygroundAttachment(attachment); err != nil {
		_ = s.driver.Delete(ctx, objectKey)
		return nil, err
	}
	return attachment, nil
}

func (s *Service) List(ctx context.Context, userID int, sessionID string) ([]*model.PlaygroundAttachment, error) {
	settings := normalizedSettings(s.settingsProvider())
	if !settings.Enabled {
		return nil, attachmentError(ErrCodeFeatureDisabled)
	}
	if userID <= 0 || strings.TrimSpace(sessionID) == "" {
		return nil, attachmentError(ErrCodeInvalidRequest)
	}
	return model.ListPlaygroundAttachmentsForSession(userID, sessionID, s.now().Unix())
}

func (s *Service) Get(ctx context.Context, userID int, attachmentID string) (*model.PlaygroundAttachment, error) {
	settings := normalizedSettings(s.settingsProvider())
	if !settings.Enabled {
		return nil, attachmentError(ErrCodeFeatureDisabled)
	}
	attachment, err := model.GetPlaygroundAttachment(attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, attachmentError(ErrCodeAttachmentNotFound)
	}
	if attachment.UserID != userID {
		return nil, attachmentError(ErrCodeAccessDenied)
	}
	return attachment, nil
}

func (s *Service) Delete(ctx context.Context, userID int, attachmentID string) error {
	attachment, err := s.Get(ctx, userID, attachmentID)
	if err != nil {
		return err
	}
	if attachment.Status != model.PlaygroundAttachmentStatusActive {
		return nil
	}
	storageDriver, err := s.resolveAttachmentDriver(attachment)
	if err != nil {
		return err
	}
	if err := storageDriver.Delete(ctx, attachment.ObjectKey); err != nil {
		return attachmentError(ErrCodeStorageDeleteFailed)
	}
	deleted, err := model.MarkPlaygroundAttachmentDeleted(attachment.ID, userID, s.now().Unix())
	if err != nil {
		return err
	}
	if !deleted {
		return attachmentError(ErrCodeAttachmentNotFound)
	}
	return nil
}

func (s *Service) GenerateReference(ctx context.Context, input ReferenceInput) (*Reference, error) {
	references, err := s.GenerateReferences(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(references) == 0 {
		return nil, attachmentError(ErrCodeAttachmentNotFound)
	}
	return &references[0], nil
}

func (s *Service) GenerateReferences(ctx context.Context, input ReferenceInput) ([]Reference, error) {
	settings := normalizedSettings(s.settingsProvider())
	if !settings.Enabled {
		return nil, attachmentError(ErrCodeFeatureDisabled)
	}
	if input.UserID <= 0 || len(input.AttachmentIDs) == 0 {
		return nil, attachmentError(ErrCodeInvalidRequest)
	}
	if settings.MaxFilesPerMessage > 0 && len(input.AttachmentIDs) > settings.MaxFilesPerMessage {
		return nil, attachmentError(ErrCodeTooManyAttachments)
	}

	now := s.now()
	expiresAt := now.Add(time.Duration(settings.ReferenceTTLSeconds) * time.Second)
	references := make([]Reference, 0, len(input.AttachmentIDs))
	for _, attachmentID := range input.AttachmentIDs {
		attachment, err := model.GetPlaygroundAttachment(attachmentID)
		if err != nil {
			return nil, err
		}
		if attachment == nil {
			return nil, attachmentError(ErrCodeAttachmentNotFound)
		}
		if attachment.UserID != input.UserID {
			return nil, attachmentError(ErrCodeAccessDenied)
		}
		if attachment.Status != model.PlaygroundAttachmentStatusActive || attachment.ExpiresAt <= now.Unix() {
			return nil, attachmentError(ErrCodeAttachmentExpired)
		}

		reference, err := s.referenceForAttachment(ctx, attachment, input.PublicBaseURL, expiresAt)
		if err != nil {
			return nil, err
		}
		if err := model.TouchPlaygroundAttachmentLastUsed(attachment.ID, now.Unix()); err != nil {
			return nil, err
		}
		references = append(references, *reference)
	}
	return references, nil
}

func (s *Service) VerifyLocalReference(ctx context.Context, input LocalReferenceInput) (*model.PlaygroundAttachment, error) {
	now := s.now().Unix()
	if input.AttachmentID == "" || input.ObjectKey == "" || input.ExpiresAt <= now || input.Signature == "" {
		return nil, attachmentError(ErrCodeInvalidSignedReference)
	}
	expectedSignature := signLocalReference(input.AttachmentID, input.ObjectKey, input.ExpiresAt)
	if !hmac.Equal([]byte(input.Signature), []byte(expectedSignature)) {
		return nil, attachmentError(ErrCodeInvalidSignedReference)
	}
	attachment, err := model.GetPlaygroundAttachment(input.AttachmentID)
	if err != nil {
		return nil, err
	}
	if attachment == nil ||
		attachment.Driver != "local" ||
		attachment.ObjectKey != input.ObjectKey ||
		attachment.Status != model.PlaygroundAttachmentStatusActive ||
		attachment.ExpiresAt <= now {
		return nil, attachmentError(ErrCodeInvalidSignedReference)
	}
	if err := model.TouchPlaygroundAttachmentLastUsed(attachment.ID, now); err != nil {
		return nil, err
	}
	return attachment, nil
}

func (s *Service) Open(ctx context.Context, attachment *model.PlaygroundAttachment) (io.ReadCloser, error) {
	storageDriver, err := s.resolveAttachmentDriver(attachment)
	if err != nil {
		return nil, err
	}
	reader, err := storageDriver.Open(ctx, attachment.ObjectKey)
	if err != nil {
		return nil, attachmentError(ErrCodeStorageReadFailed)
	}
	return reader, nil
}

func (s *Service) CleanupExpired(ctx context.Context, limit int) (CleanupResult, error) {
	now := s.now().Unix()
	attachments, err := model.FindExpiredPlaygroundAttachments(now, limit)
	if err != nil {
		return CleanupResult{}, err
	}
	result := CleanupResult{}
	for _, attachment := range attachments {
		storageDriver, err := s.resolveAttachmentDriver(attachment)
		if err != nil {
			result.FailedCount++
			continue
		}
		if err := storageDriver.Delete(ctx, attachment.ObjectKey); err != nil && !errors.Is(err, os.ErrNotExist) {
			result.FailedCount++
			continue
		}
		if err := model.MarkPlaygroundAttachmentExpired(attachment.ID, now); err != nil {
			return result, err
		}
		result.ExpiredCount++
	}
	return result, nil
}

func (s *Service) referenceForAttachment(ctx context.Context, attachment *model.PlaygroundAttachment, publicBaseURL string, expiresAt time.Time) (*Reference, error) {
	storageDriver, err := s.resolveAttachmentDriver(attachment)
	if err != nil {
		return nil, err
	}
	storageReference, err := storageDriver.Reference(ctx, attachment.ObjectKey, expiresAt)
	if err == nil && storageReference != nil {
		return &Reference{
			ID:        attachment.ID,
			SessionID: attachment.SessionID,
			URL:       storageReference.URL,
			ExpiresAt: storageReference.ExpiresAt,
			Filename:  attachment.Filename,
			MimeType:  attachment.MimeType,
			Size:      attachment.Size,
			Status:    attachment.Status,
		}, nil
	}
	if err != nil && err != driver.ErrReferenceUnsupported {
		return nil, attachmentError(ErrCodeStorageDriverUnavailable)
	}
	if attachment.Driver != "local" {
		return nil, attachmentError(ErrCodeStorageDriverUnavailable)
	}
	expires := expiresAt.Unix()
	signature := signLocalReference(attachment.ID, attachment.ObjectKey, expires)
	referenceURL, err := localReferenceURL(publicBaseURL, attachment.ID, attachment.ObjectKey, expires, signature)
	if err != nil {
		return nil, err
	}
	return &Reference{
		ID:        attachment.ID,
		SessionID: attachment.SessionID,
		URL:       referenceURL,
		ExpiresAt: expires,
		Signature: signature,
		Filename:  attachment.Filename,
		MimeType:  attachment.MimeType,
		Size:      attachment.Size,
		Status:    attachment.Status,
	}, nil
}

func normalizedSettings(settings Settings) Settings {
	if settings.StorageDriver == "" {
		settings.StorageDriver = "local"
	}
	if settings.TTLHours <= 0 {
		settings.TTLHours = 24
	}
	if settings.MaxFileSizeBytes <= 0 {
		settings.MaxFileSizeBytes = 10 * 1024 * 1024
	}
	if settings.MaxFilesPerMessage <= 0 {
		settings.MaxFilesPerMessage = 4
	}
	if settings.MaxFilesPerSession <= 0 {
		settings.MaxFilesPerSession = 20
	}
	if settings.ReferenceTTLSeconds <= 0 {
		settings.ReferenceTTLSeconds = 300
	}
	return settings
}

func readUploadBytes(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, attachmentError(ErrCodeFileTooLarge)
	}
	if len(data) == 0 {
		return nil, attachmentError(ErrCodeInvalidRequest)
	}
	return data, nil
}

func detectUploadMIME(data []byte, filename string, declared string) string {
	detected := normalizeMIME(http.DetectContentType(data))
	extensionMIME := normalizeMIME(mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))))
	declaredMIME := normalizeMIME(declared)

	if detected == "application/octet-stream" || detected == "application/zip" {
		if extensionMIME != "" {
			return extensionMIME
		}
		if declaredMIME != "" {
			return declaredMIME
		}
	}
	return detected
}

func normalizeMIME(value string) string {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return strings.ToLower(strings.TrimSpace(value))
	}
	return strings.ToLower(mediaType)
}

func isMIMEAllowed(mimeType string, allowed []string) bool {
	for _, allowedType := range allowed {
		allowedType = normalizeMIME(allowedType)
		if allowedType == mimeType {
			return true
		}
		if strings.HasSuffix(allowedType, "/*") && strings.HasPrefix(mimeType, strings.TrimSuffix(allowedType, "*")) {
			return true
		}
	}
	return false
}

func isDecodableImage(mimeType string, data []byte) bool {
	if mimeType == "image/svg+xml" {
		return true
	}
	_, _, err := image.DecodeConfig(bytes.NewReader(data))
	return err == nil
}

func sanitizeFilename(filename string) string {
	base := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	base = strings.TrimSpace(base)
	if base == "." || base == "/" || base == "" {
		return "attachment"
	}
	if len(base) > 255 {
		return base[:255]
	}
	return base
}

func newAttachmentID() (string, error) {
	random, err := common.GenerateRandomCharsKey(32)
	if err != nil {
		return "", err
	}
	return "att_" + random, nil
}

func newObjectKey(userID int, attachmentID string, now time.Time) string {
	return fmt.Sprintf("users/%d/%s/%s", userID, now.UTC().Format("20060102"), attachmentID)
}

func signLocalReference(attachmentID string, objectKey string, expiresAt int64) string {
	message := attachmentID + "\n" + objectKey + "\n" + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, []byte(common.SessionSecret))
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func localReferenceURL(publicBaseURL string, attachmentID string, objectKey string, expiresAt int64, signature string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		base = "/"
	}
	referencePath := "/pg/attachments/" + url.PathEscape(attachmentID) + "/content"
	if base == "/" {
		base = ""
	}
	parsed, err := url.Parse(base + referencePath)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("key", objectKey)
	query.Set("expires", strconv.FormatInt(expiresAt, 10))
	query.Set("signature", signature)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func attachmentError(code string) *AttachmentError {
	return &AttachmentError{Code: code, Message: code}
}

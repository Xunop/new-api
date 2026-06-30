package playground_attachment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/playground_attachment/driver"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeDriver struct {
	name       string
	writes     map[string][]byte
	deletes    []string
	deleteErr  error
	references map[string]string
}

func newFakeDriver(name string) *fakeDriver {
	return &fakeDriver{
		name:       name,
		writes:     map[string][]byte{},
		references: map[string]string{},
	}
}

func (d *fakeDriver) Name() string {
	return d.name
}

func (d *fakeDriver) Put(ctx context.Context, objectKey string, contentType string, body io.Reader) (int64, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, err
	}
	d.writes[objectKey] = data
	return int64(len(data)), nil
}

func (d *fakeDriver) Open(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	data, ok := d.writes[objectKey]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (d *fakeDriver) Delete(ctx context.Context, objectKey string) error {
	d.deletes = append(d.deletes, objectKey)
	if d.deleteErr != nil {
		return d.deleteErr
	}
	delete(d.writes, objectKey)
	return nil
}

func (d *fakeDriver) Reference(ctx context.Context, objectKey string, expiresAt time.Time) (*driver.Reference, error) {
	if ref, ok := d.references[objectKey]; ok {
		return &driver.Reference{URL: ref, ExpiresAt: expiresAt.Unix()}, nil
	}
	return nil, driver.ErrReferenceUnsupported
}

func setupAttachmentServiceTest(t *testing.T) {
	t.Helper()

	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.PlaygroundAttachment{}))

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func testSettings() Settings {
	return Settings{
		Enabled:             true,
		StorageDriver:       "local",
		TTLHours:            24,
		MaxFileSizeBytes:    1024,
		MaxFilesPerMessage:  4,
		MaxFilesPerSession:  8,
		AllowedMIMETypes:    []string{"image/png", "text/plain", "application/pdf"},
		ReferenceTTLSeconds: 300,
	}
}

func newTestService(d driver.Driver, settings Settings, now time.Time) *Service {
	svc := NewService(d, func() Settings { return settings })
	svc.now = func() time.Time { return now }
	return svc
}

func requireAttachmentErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var attachmentErr *AttachmentError
	require.ErrorAs(t, err, &attachmentErr)
	assert.Equal(t, code, attachmentErr.Code)
}

func TestUploadStoresAllowedFileMetadataAndBody(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	now := time.Unix(1700000000, 0)
	svc := newTestService(storage, testSettings(), now)

	attachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:      7,
		SessionID:   "session-a",
		Filename:    "../report.png",
		ContentType: "image/png",
		Reader:      bytes.NewReader(testPNGBytes()),
	})

	require.NoError(t, err)
	require.NotEmpty(t, attachment.ID)
	assert.Equal(t, 7, attachment.UserID)
	assert.Equal(t, "session-a", attachment.SessionID)
	assert.Equal(t, "report.png", attachment.Filename)
	assert.Equal(t, "image/png", attachment.MimeType)
	assert.Equal(t, model.PlaygroundAttachmentStatusActive, attachment.Status)
	assert.Equal(t, now.Add(24*time.Hour).Unix(), attachment.ExpiresAt)
	assert.NotContains(t, attachment.ObjectKey, "report.png")
	assert.NotEmpty(t, storage.writes[attachment.ObjectKey])

	var persisted model.PlaygroundAttachment
	require.NoError(t, model.DB.First(&persisted, "id = ?", attachment.ID).Error)
	assert.Equal(t, attachment.Digest, persisted.Digest)
}

func TestUploadRejectsOversizedFileBeforeStorageWrite(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	settings := testSettings()
	settings.MaxFileSizeBytes = 3
	svc := newTestService(storage, settings, time.Unix(1700000000, 0))

	_, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("four")),
	})

	requireAttachmentErrorCode(t, err, ErrCodeFileTooLarge)
	assert.Empty(t, storage.writes)
}

func TestUploadRejectsDisallowedMIME(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	settings := testSettings()
	settings.AllowedMIMETypes = []string{"image/png"}
	svc := newTestService(storage, settings, time.Unix(1700000000, 0))

	_, err := svc.Upload(context.Background(), UploadInput{
		UserID:      7,
		SessionID:   "session-a",
		Filename:    "note.txt",
		ContentType: "text/plain",
		Reader:      bytes.NewReader([]byte("plain text")),
	})

	requireAttachmentErrorCode(t, err, ErrCodeMIMENotAllowed)
	assert.Empty(t, storage.writes)
}

func TestUploadRejectsWhenSessionAttachmentLimitReached(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	settings := testSettings()
	settings.MaxFilesPerSession = 1
	svc := newTestService(storage, settings, time.Unix(1700000000, 0))

	_, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-1.txt",
		Reader:    bytes.NewReader([]byte("first")),
	})
	require.NoError(t, err)

	_, err = svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-2.txt",
		Reader:    bytes.NewReader([]byte("second")),
	})

	requireAttachmentErrorCode(t, err, ErrCodeTooManyAttachments)
}

func TestUploadRejectsWhenMessageAttachmentLimitReached(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	settings := testSettings()
	settings.MaxFilesPerMessage = 1
	settings.MaxFilesPerSession = 8
	svc := newTestService(storage, settings, time.Unix(1700000000, 0))

	_, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-1.txt",
		Reader:    bytes.NewReader([]byte("first")),
	})
	require.NoError(t, err)

	_, err = svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-2.txt",
		Reader:    bytes.NewReader([]byte("second")),
	})

	requireAttachmentErrorCode(t, err, ErrCodeTooManyAttachments)
}

func TestUploadRejectsUnknownStorageDriver(t *testing.T) {
	setupAttachmentServiceTest(t)

	settings := testSettings()
	settings.StorageDriver = "bogus"
	svc := NewService(newConfiguredDriver(settings), func() Settings { return settings })
	svc.now = func() time.Time { return time.Unix(1700000000, 0) }

	_, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("plain text")),
	})

	requireAttachmentErrorCode(t, err, ErrCodeStorageDriverUnavailable)
}

func TestListReturnsMetadataWhenConfiguredDriverIsUnknown(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	now := time.Unix(1700000000, 0)
	uploadSvc := newTestService(storage, testSettings(), now)
	attachment, err := uploadSvc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("plain text")),
	})
	require.NoError(t, err)

	settings := testSettings()
	settings.StorageDriver = "bogus"
	listSvc := NewService(newConfiguredDriver(settings), func() Settings { return settings })
	listSvc.now = func() time.Time { return now }

	attachments, err := listSvc.List(context.Background(), 7, "session-a")

	require.NoError(t, err)
	require.Len(t, attachments, 1)
	assert.Equal(t, attachment.ID, attachments[0].ID)
}

func TestGenerateReferencesRejectsOtherUsersAttachment(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	attachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("plain text")),
	})
	require.NoError(t, err)

	_, err = svc.GenerateReferences(context.Background(), ReferenceInput{
		UserID:        8,
		AttachmentIDs: []string{attachment.ID},
		PublicBaseURL: "https://gateway.example",
	})

	requireAttachmentErrorCode(t, err, ErrCodeAccessDenied)
}

func TestGenerateReferencesRejectsWhenMessageAttachmentLimitExceeded(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	settings := testSettings()
	settings.MaxFilesPerMessage = 2
	svc := newTestService(storage, settings, time.Unix(1700000000, 0))

	firstAttachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-1.txt",
		Reader:    bytes.NewReader([]byte("first")),
	})
	require.NoError(t, err)

	secondAttachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note-2.txt",
		Reader:    bytes.NewReader([]byte("second")),
	})
	require.NoError(t, err)

	settings.MaxFilesPerMessage = 1
	svc = newTestService(storage, settings, time.Unix(1700000000, 0))

	_, err = svc.GenerateReferences(context.Background(), ReferenceInput{
		UserID:        7,
		AttachmentIDs: []string{firstAttachment.ID, secondAttachment.ID},
		PublicBaseURL: "https://gateway.example",
	})

	requireAttachmentErrorCode(t, err, ErrCodeTooManyAttachments)
}

func TestGetAndDeleteRejectOtherUsersAttachment(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	attachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("plain text")),
	})
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), 8, attachment.ID)
	requireAttachmentErrorCode(t, err, ErrCodeAccessDenied)

	err = svc.Delete(context.Background(), 8, attachment.ID)
	requireAttachmentErrorCode(t, err, ErrCodeAccessDenied)
	assert.Empty(t, storage.deletes)
}

func TestGenerateReferencesRejectsExpiredAttachment(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	expired := &model.PlaygroundAttachment{
		ID:        "att_expired",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "local",
		ObjectKey: "objects/expired",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1699999999, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(expired).Error)

	_, err := svc.GenerateReferences(context.Background(), ReferenceInput{
		UserID:        7,
		AttachmentIDs: []string{expired.ID},
		PublicBaseURL: "https://gateway.example",
	})

	requireAttachmentErrorCode(t, err, ErrCodeAttachmentExpired)
}

func TestGenerateReferencesRejectsUnknownPersistedDriver(t *testing.T) {
	setupAttachmentServiceTest(t)

	svc := newTestService(newFakeDriver("local"), testSettings(), time.Unix(1700000000, 0))
	attachment := &model.PlaygroundAttachment{
		ID:        "att_unknown_driver",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "bogus",
		ObjectKey: "objects/bogus",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1700003600, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(attachment).Error)

	var err error
	require.NotPanics(t, func() {
		_, err = svc.GenerateReferences(context.Background(), ReferenceInput{
			UserID:        7,
			AttachmentIDs: []string{attachment.ID},
			PublicBaseURL: "https://gateway.example",
		})
	})
	requireAttachmentErrorCode(t, err, ErrCodeStorageDriverUnavailable)
}

func TestCleanupExpiredDeletesObjectsAndMarksMetadataExpired(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	activeExpired := &model.PlaygroundAttachment{
		ID:        "att_cleanup",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "local",
		ObjectKey: "objects/cleanup",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1699999999, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(activeExpired).Error)

	result, err := svc.CleanupExpired(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ExpiredCount)
	assert.Equal(t, []string{"objects/cleanup"}, storage.deletes)

	var reloaded model.PlaygroundAttachment
	require.NoError(t, model.DB.First(&reloaded, "id = ?", activeExpired.ID).Error)
	assert.Equal(t, model.PlaygroundAttachmentStatusExpired, reloaded.Status)
}

func TestCleanupExpiredKeepsMetadataActiveWhenDeleteFails(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	storage.deleteErr = errors.New("storage unavailable")
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	activeExpired := &model.PlaygroundAttachment{
		ID:        "att_cleanup_retry",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "local",
		ObjectKey: "objects/retry",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1699999999, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(activeExpired).Error)

	result, err := svc.CleanupExpired(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExpiredCount)
	assert.Equal(t, 1, result.FailedCount)

	var reloaded model.PlaygroundAttachment
	require.NoError(t, model.DB.First(&reloaded, "id = ?", activeExpired.ID).Error)
	assert.Equal(t, model.PlaygroundAttachmentStatusActive, reloaded.Status)
}

func TestCleanupExpiredMarksMetadataExpiredWhenObjectAlreadyMissing(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	storage.deleteErr = os.ErrNotExist
	svc := newTestService(storage, testSettings(), time.Unix(1700000000, 0))
	activeExpired := &model.PlaygroundAttachment{
		ID:        "att_cleanup_missing_object",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "local",
		ObjectKey: "objects/missing",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1699999999, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(activeExpired).Error)

	result, err := svc.CleanupExpired(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ExpiredCount)
	assert.Equal(t, 0, result.FailedCount)

	var reloaded model.PlaygroundAttachment
	require.NoError(t, model.DB.First(&reloaded, "id = ?", activeExpired.ID).Error)
	assert.Equal(t, model.PlaygroundAttachmentStatusExpired, reloaded.Status)
}

func TestCleanupExpiredCountsUnknownDriverAsFailure(t *testing.T) {
	setupAttachmentServiceTest(t)

	svc := newTestService(newFakeDriver("local"), testSettings(), time.Unix(1700000000, 0))
	activeExpired := &model.PlaygroundAttachment{
		ID:        "att_cleanup_unknown_driver",
		UserID:    7,
		SessionID: "session-a",
		Driver:    "bogus",
		ObjectKey: "objects/retry",
		Filename:  "note.txt",
		MimeType:  "text/plain",
		Size:      5,
		Digest:    "digest",
		Status:    model.PlaygroundAttachmentStatusActive,
		CreatedAt: time.Unix(1699990000, 0).Unix(),
		ExpiresAt: time.Unix(1699999999, 0).Unix(),
	}
	require.NoError(t, model.DB.Create(activeExpired).Error)

	var (
		result CleanupResult
		err    error
	)
	require.NotPanics(t, func() {
		result, err = svc.CleanupExpired(context.Background(), 10)
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExpiredCount)
	assert.Equal(t, 1, result.FailedCount)

	var reloaded model.PlaygroundAttachment
	require.NoError(t, model.DB.First(&reloaded, "id = ?", activeExpired.ID).Error)
	assert.Equal(t, model.PlaygroundAttachmentStatusActive, reloaded.Status)
}

func TestVerifyLocalReferenceRejectsBadAndExpiredSignatures(t *testing.T) {
	setupAttachmentServiceTest(t)

	storage := newFakeDriver("local")
	now := time.Unix(1700000000, 0)
	svc := newTestService(storage, testSettings(), now)
	attachment, err := svc.Upload(context.Background(), UploadInput{
		UserID:    7,
		SessionID: "session-a",
		Filename:  "note.txt",
		Reader:    bytes.NewReader([]byte("plain text")),
	})
	require.NoError(t, err)
	reference, err := svc.GenerateReference(context.Background(), ReferenceInput{
		UserID:        7,
		AttachmentIDs: []string{attachment.ID},
		PublicBaseURL: "https://gateway.example",
	})
	require.NoError(t, err)

	_, err = svc.VerifyLocalReference(context.Background(), LocalReferenceInput{
		AttachmentID: attachment.ID,
		ObjectKey:    attachment.ObjectKey,
		ExpiresAt:    reference.ExpiresAt,
		Signature:    "bad-signature",
	})
	requireAttachmentErrorCode(t, err, ErrCodeInvalidSignedReference)

	svc.now = func() time.Time { return time.Unix(reference.ExpiresAt+1, 0) }
	_, err = svc.VerifyLocalReference(context.Background(), LocalReferenceInput{
		AttachmentID: attachment.ID,
		ObjectKey:    attachment.ObjectKey,
		ExpiresAt:    reference.ExpiresAt,
		Signature:    reference.Signature,
	})
	requireAttachmentErrorCode(t, err, ErrCodeInvalidSignedReference)
}

func TestLocalDriverRejectsOriginalFilenamePathTraversal(t *testing.T) {
	baseDir := t.TempDir()
	localDriver := NewLocalDriver(baseDir)

	_, err := localDriver.Put(context.Background(), "../evil.txt", "text/plain", bytes.NewReader([]byte("x")))
	require.Error(t, err)

	written, err := localDriver.Put(context.Background(), "users/7/att_safe", "text/plain", bytes.NewReader([]byte("safe")))
	require.NoError(t, err)
	assert.EqualValues(t, 4, written)

	_, err = os.Stat(filepath.Join(baseDir, "users", "7", "att_safe"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(baseDir, "evil.txt"))
	assert.True(t, os.IsNotExist(err))
}

func testPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xdd, 0x8d,
		0xb0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

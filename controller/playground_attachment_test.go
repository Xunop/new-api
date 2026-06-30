package controller

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	playgroundattachment "github.com/QuantumNous/new-api/service/playground_attachment"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type playgroundAttachmentAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Data    any    `json:"data"`
}

func setupPlaygroundAttachmentControllerTest(t *testing.T) {
	t.Helper()

	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.PlaygroundAttachment{}))

	settings := system_setting.GetPlaygroundAttachmentSettings()
	original := *settings
	settings.Enabled = true
	settings.StorageDriver = "local"
	settings.LocalBasePath = t.TempDir()
	settings.AllowedMIMETypes = []string{"text/plain"}
	settings.MaxFileSizeBytes = 1024
	settings.MaxFilesPerMessage = 4
	settings.MaxFilesPerSession = 8
	settings.ReferenceTTLSeconds = 300
	settings.TTLHours = 24
	t.Cleanup(func() {
		*settings = original
	})
}

func newMultipartUploadRequest(t *testing.T, filename string, content string, sessionID string) (*http.Request, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("session_id", sessionID))
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/pg/attachments", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, writer.FormDataContentType()
}

func decodePlaygroundAttachmentAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) playgroundAttachmentAPIResponse {
	t.Helper()
	var payload playgroundAttachmentAPIResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	return payload
}

func uploadPlaygroundAttachmentForControllerTest(t *testing.T, userID int) map[string]any {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", userID)
	request, _ := newMultipartUploadRequest(t, "notes.txt", "plain text", "session-a")
	ctx.Request = request

	UploadPlaygroundAttachment(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	require.True(t, payload.Success, payload.Message)
	data, ok := payload.Data.(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, data["id"])
	return data
}

func TestUploadPlaygroundAttachmentReturnsMetadata(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)

	data := uploadPlaygroundAttachmentForControllerTest(t, 9)

	assert.Equal(t, "notes.txt", data["filename"])
	assert.Equal(t, "text/plain", data["mime_type"])
	assert.Equal(t, "active", data["status"])
}

func TestUploadPlaygroundAttachmentRejectsOversizedMultipartBodyWithStableCode(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 9)
	request, _ := newMultipartUploadRequest(t, "notes.txt", strings.Repeat("x", 2<<20), "session-a")
	ctx.Request = request

	UploadPlaygroundAttachment(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	assert.False(t, payload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeFileTooLarge, payload.Code)
}

func TestUploadPlaygroundAttachmentRejectsWhenAttachmentLimitReached(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)

	settings := system_setting.GetPlaygroundAttachmentSettings()
	settings.MaxFilesPerMessage = 1
	settings.MaxFilesPerSession = 8

	firstRecorder := httptest.NewRecorder()
	firstCtx, _ := gin.CreateTestContext(firstRecorder)
	firstCtx.Set("id", 9)
	firstRequest, _ := newMultipartUploadRequest(t, "first.txt", "plain text", "session-a")
	firstCtx.Request = firstRequest

	UploadPlaygroundAttachment(firstCtx)

	require.Equal(t, http.StatusOK, firstRecorder.Code)
	firstPayload := decodePlaygroundAttachmentAPIResponse(t, firstRecorder)
	require.True(t, firstPayload.Success, firstPayload.Message)

	secondRecorder := httptest.NewRecorder()
	secondCtx, _ := gin.CreateTestContext(secondRecorder)
	secondCtx.Set("id", 9)
	secondRequest, _ := newMultipartUploadRequest(t, "second.txt", "plain text", "session-a")
	secondCtx.Request = secondRequest

	UploadPlaygroundAttachment(secondCtx)

	require.Equal(t, http.StatusOK, secondRecorder.Code)
	secondPayload := decodePlaygroundAttachmentAPIResponse(t, secondRecorder)
	assert.False(t, secondPayload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeTooManyAttachments, secondPayload.Code)
}

func TestGeneratePlaygroundAttachmentReferencesReturnsSignedURL(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 9)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/pg/attachments/references",
		strings.NewReader(`{"attachment_ids":["`+attachmentID+`"]}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	GeneratePlaygroundAttachmentReferences(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	require.True(t, payload.Success, payload.Message)
	references, ok := payload.Data.([]any)
	require.True(t, ok)
	require.Len(t, references, 1)
	reference := references[0].(map[string]any)
	assert.Equal(t, attachmentID, reference["id"])
	assert.Contains(t, reference["url"], "/pg/attachments/"+attachmentID+"/content")
	assert.NotEmpty(t, reference["signature"])
}

func TestListAndGetPlaygroundAttachmentsScopeByUserAndSession(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)

	firstAttachment := uploadPlaygroundAttachmentForControllerTest(t, 9)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 9)
	secondRequest, _ := newMultipartUploadRequest(t, "other.txt", "other text", "session-b")
	ctx.Request = secondRequest
	UploadPlaygroundAttachment(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Set("id", 9)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/pg/attachments?session_id=session-a", nil)
	listCtx.Request.URL.RawQuery = "session_id=session-a"

	ListPlaygroundAttachments(listCtx)

	require.Equal(t, http.StatusOK, listRecorder.Code)
	listPayload := decodePlaygroundAttachmentAPIResponse(t, listRecorder)
	require.True(t, listPayload.Success, listPayload.Message)
	attachments, ok := listPayload.Data.([]any)
	require.True(t, ok)
	require.Len(t, attachments, 1)
	listedAttachment := attachments[0].(map[string]any)
	assert.Equal(t, firstAttachment["id"], listedAttachment["id"])

	getRecorder := httptest.NewRecorder()
	getCtx, _ := gin.CreateTestContext(getRecorder)
	getCtx.Set("id", 9)
	getCtx.Params = gin.Params{{Key: "id", Value: firstAttachment["id"].(string)}}
	getCtx.Request = httptest.NewRequest(http.MethodGet, "/pg/attachments/"+firstAttachment["id"].(string), nil)

	GetPlaygroundAttachment(getCtx)

	require.Equal(t, http.StatusOK, getRecorder.Code)
	getPayload := decodePlaygroundAttachmentAPIResponse(t, getRecorder)
	require.True(t, getPayload.Success, getPayload.Message)
	getData, ok := getPayload.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, firstAttachment["id"], getData["id"])

	otherUserRecorder := httptest.NewRecorder()
	otherUserCtx, _ := gin.CreateTestContext(otherUserRecorder)
	otherUserCtx.Set("id", 10)
	otherUserCtx.Params = gin.Params{{Key: "id", Value: firstAttachment["id"].(string)}}
	otherUserCtx.Request = httptest.NewRequest(http.MethodGet, "/pg/attachments/"+firstAttachment["id"].(string), nil)

	GetPlaygroundAttachment(otherUserCtx)

	require.Equal(t, http.StatusOK, otherUserRecorder.Code)
	otherUserPayload := decodePlaygroundAttachmentAPIResponse(t, otherUserRecorder)
	assert.False(t, otherUserPayload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeAccessDenied, otherUserPayload.Code)
}

func TestDeletePlaygroundAttachmentInvalidatesFutureReferences(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	deleteRecorder := httptest.NewRecorder()
	deleteCtx, _ := gin.CreateTestContext(deleteRecorder)
	deleteCtx.Set("id", 9)
	deleteCtx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	deleteCtx.Request = httptest.NewRequest(http.MethodDelete, "/pg/attachments/"+attachmentID, nil)

	DeletePlaygroundAttachment(deleteCtx)

	require.Equal(t, http.StatusOK, deleteRecorder.Code)
	deletePayload := decodePlaygroundAttachmentAPIResponse(t, deleteRecorder)
	require.True(t, deletePayload.Success, deletePayload.Message)

	referenceRecorder := httptest.NewRecorder()
	referenceCtx, _ := gin.CreateTestContext(referenceRecorder)
	referenceCtx.Set("id", 9)
	referenceCtx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	referenceCtx.Request = httptest.NewRequest(http.MethodPost, "/pg/attachments/"+attachmentID+"/reference", nil)

	GeneratePlaygroundAttachmentReference(referenceCtx)

	require.Equal(t, http.StatusOK, referenceRecorder.Code)
	referencePayload := decodePlaygroundAttachmentAPIResponse(t, referenceRecorder)
	assert.False(t, referencePayload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeAttachmentExpired, referencePayload.Code)
}

func TestGeneratePlaygroundAttachmentReferencesRejectsTooManyAttachments(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)

	settings := system_setting.GetPlaygroundAttachmentSettings()
	settings.MaxFilesPerMessage = 8

	attachmentIDs := make([]string, 0, 5)
	for index := 0; index < 5; index++ {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", 9)
		request, _ := newMultipartUploadRequest(t, "notes.txt", "plain text", "session-a")
		ctx.Request = request
		UploadPlaygroundAttachment(ctx)

		require.Equal(t, http.StatusOK, recorder.Code)
		payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
		require.True(t, payload.Success, payload.Message)
		data, ok := payload.Data.(map[string]any)
		require.True(t, ok)
		attachmentIDs = append(attachmentIDs, data["id"].(string))
	}

	settings.MaxFilesPerMessage = 4

	body := `{"attachment_ids":["` + strings.Join(attachmentIDs, `","`) + `"]}`
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 9)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/pg/attachments/references",
		strings.NewReader(body),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	GeneratePlaygroundAttachmentReferences(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	assert.False(t, payload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeTooManyAttachments, payload.Code)
}

func TestReadPlaygroundAttachmentContentUsesSignedReferenceOnly(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	svc := playgroundattachment.NewDefaultService()
	reference, err := svc.GenerateReference(t.Context(), playgroundattachment.ReferenceInput{
		UserID:        9,
		AttachmentIDs: []string{attachmentID},
		PublicBaseURL: "https://gateway.example",
	})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	request := httptest.NewRequest(http.MethodGet, reference.URL, nil)
	ctx.Request = request

	ReadPlaygroundAttachmentContent(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "plain text", recorder.Body.String())
}

func TestReadPlaygroundAttachmentContentRejectsBadSignature(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/pg/attachments/"+attachmentID+"/content?key=bad&expires=9999999999&signature=bad",
		nil,
	)

	ReadPlaygroundAttachmentContent(ctx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	assert.False(t, payload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeInvalidSignedReference, payload.Code)
}

func TestReadPlaygroundAttachmentContentRejectsDeletedAttachment(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	svc := playgroundattachment.NewDefaultService()
	reference, err := svc.GenerateReference(t.Context(), playgroundattachment.ReferenceInput{
		UserID:        9,
		AttachmentIDs: []string{attachmentID},
		PublicBaseURL: "https://gateway.example",
	})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(t.Context(), 9, attachmentID))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	ctx.Request = httptest.NewRequest(http.MethodGet, reference.URL, nil)

	ReadPlaygroundAttachmentContent(ctx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	assert.False(t, payload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeInvalidSignedReference, payload.Code)
}

func TestReadPlaygroundAttachmentContentRejectsMismatchedObjectKey(t *testing.T) {
	setupPlaygroundAttachmentControllerTest(t)
	data := uploadPlaygroundAttachmentForControllerTest(t, 9)
	attachmentID := data["id"].(string)

	svc := playgroundattachment.NewDefaultService()
	reference, err := svc.GenerateReference(t.Context(), playgroundattachment.ReferenceInput{
		UserID:        9,
		AttachmentIDs: []string{attachmentID},
		PublicBaseURL: "https://gateway.example",
	})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: attachmentID}}
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		strings.Replace(reference.URL, "key=users%2F9%2F", "key=users%2F10%2F", 1),
		nil,
	)

	ReadPlaygroundAttachmentContent(ctx)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	payload := decodePlaygroundAttachmentAPIResponse(t, recorder)
	assert.False(t, payload.Success)
	assert.Equal(t, playgroundattachment.ErrCodeInvalidSignedReference, payload.Code)
}

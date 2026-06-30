package controller

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	playgroundattachment "github.com/QuantumNous/new-api/service/playground_attachment"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

type playgroundAttachmentReferenceRequest struct {
	AttachmentIDs []string `json:"attachment_ids"`
}

func UploadPlaygroundAttachment(c *gin.Context) {
	settings := system_setting.GetPlaygroundAttachmentSettings()
	if settings.MaxFileSizeBytes > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, settings.MaxFileSizeBytes+(1<<20))
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		code := playgroundattachment.ErrCodeInvalidRequest
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) || strings.Contains(err.Error(), "http: request body too large") {
			code = playgroundattachment.ErrCodeFileTooLarge
		}
		playgroundAttachmentError(c, http.StatusOK, code, err)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		playgroundAttachmentError(c, http.StatusOK, playgroundattachment.ErrCodeInvalidRequest, err)
		return
	}
	defer file.Close()

	attachment, err := playgroundattachment.NewDefaultService().Upload(c.Request.Context(), playgroundattachment.UploadInput{
		UserID:      c.GetInt("id"),
		SessionID:   c.PostForm("session_id"),
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Reader:      file,
	})
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, attachment)
}

func ListPlaygroundAttachments(c *gin.Context) {
	attachments, err := playgroundattachment.NewDefaultService().List(c.Request.Context(), c.GetInt("id"), c.Query("session_id"))
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, attachments)
}

func GetPlaygroundAttachment(c *gin.Context) {
	attachment, err := playgroundattachment.NewDefaultService().Get(c.Request.Context(), c.GetInt("id"), c.Param("id"))
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, attachment)
}

func DeletePlaygroundAttachment(c *gin.Context) {
	if err := playgroundattachment.NewDefaultService().Delete(c.Request.Context(), c.GetInt("id"), c.Param("id")); err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GeneratePlaygroundAttachmentReference(c *gin.Context) {
	reference, err := playgroundattachment.NewDefaultService().GenerateReference(c.Request.Context(), playgroundattachment.ReferenceInput{
		UserID:        c.GetInt("id"),
		AttachmentIDs: []string{c.Param("id")},
		PublicBaseURL: requestPublicBaseURL(c),
	})
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, reference)
}

func GeneratePlaygroundAttachmentReferences(c *gin.Context) {
	var request playgroundAttachmentReferenceRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		playgroundAttachmentError(c, http.StatusOK, playgroundattachment.ErrCodeInvalidRequest, err)
		return
	}
	references, err := playgroundattachment.NewDefaultService().GenerateReferences(c.Request.Context(), playgroundattachment.ReferenceInput{
		UserID:        c.GetInt("id"),
		AttachmentIDs: request.AttachmentIDs,
		PublicBaseURL: requestPublicBaseURL(c),
	})
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusOK, err)
		return
	}
	common.ApiSuccess(c, references)
}

func ReadPlaygroundAttachmentContent(c *gin.Context) {
	expiresAt, err := strconv.ParseInt(c.Query("expires"), 10, 64)
	if err != nil {
		playgroundAttachmentError(c, http.StatusForbidden, playgroundattachment.ErrCodeInvalidSignedReference, err)
		return
	}

	svc := playgroundattachment.NewDefaultService()
	attachment, err := svc.VerifyLocalReference(c.Request.Context(), playgroundattachment.LocalReferenceInput{
		AttachmentID: c.Param("id"),
		ObjectKey:    c.Query("key"),
		ExpiresAt:    expiresAt,
		Signature:    c.Query("signature"),
	})
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusForbidden, err)
		return
	}
	reader, err := svc.Open(c.Request.Context(), attachment)
	if err != nil {
		playgroundAttachmentServiceError(c, http.StatusInternalServerError, err)
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": attachment.Filename}))
	c.DataFromReader(http.StatusOK, attachment.Size, attachment.MimeType, reader, nil)
}

func playgroundAttachmentServiceError(c *gin.Context, status int, err error) {
	var attachmentErr *playgroundattachment.AttachmentError
	if errors.As(err, &attachmentErr) {
		playgroundAttachmentError(c, status, attachmentErr.Code, err)
		return
	}
	playgroundAttachmentError(c, status, playgroundattachment.ErrCodeInvalidRequest, err)
}

func playgroundAttachmentError(c *gin.Context, status int, code string, err error) {
	message := code
	if err != nil {
		message = err.Error()
	}
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"code":    code,
	})
}

func requestPublicBaseURL(c *gin.Context) string {
	if baseURL := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/"); baseURL != "" {
		return baseURL
	}
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}

package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	PlaygroundAttachmentStatusActive  = "active"
	PlaygroundAttachmentStatusDeleted = "deleted"
	PlaygroundAttachmentStatusExpired = "expired"
)

type PlaygroundAttachment struct {
	ID         string `json:"id" gorm:"type:varchar(64);primaryKey"`
	UserID     int    `json:"user_id" gorm:"index;index:idx_pg_attachment_user_session,priority:1"`
	SessionID  string `json:"session_id" gorm:"type:varchar(128);index;index:idx_pg_attachment_user_session,priority:2"`
	Driver     string `json:"driver" gorm:"type:varchar(32);index"`
	ObjectKey  string `json:"-" gorm:"type:varchar(512);not null"`
	Filename   string `json:"filename" gorm:"type:varchar(255)"`
	MimeType   string `json:"mime_type" gorm:"type:varchar(128);index"`
	Size       int64  `json:"size" gorm:"column:size_bytes;index"`
	Digest     string `json:"digest" gorm:"type:varchar(128)"`
	Status     string `json:"status" gorm:"type:varchar(32);index"`
	CreatedAt  int64  `json:"created_at" gorm:"bigint;index"`
	ExpiresAt  int64  `json:"expires_at" gorm:"bigint;index"`
	LastUsedAt int64  `json:"last_used_at" gorm:"bigint;index"`
}

func (PlaygroundAttachment) TableName() string {
	return "playground_attachments"
}

func (attachment *PlaygroundAttachment) BeforeCreate(_ *gorm.DB) error {
	if attachment.CreatedAt == 0 {
		attachment.CreatedAt = common.GetTimestamp()
	}
	if attachment.Status == "" {
		attachment.Status = PlaygroundAttachmentStatusActive
	}
	return nil
}

func CreatePlaygroundAttachment(attachment *PlaygroundAttachment) error {
	return DB.Create(attachment).Error
}

func GetPlaygroundAttachment(id string) (*PlaygroundAttachment, error) {
	var attachment PlaygroundAttachment
	if err := DB.Where("id = ?", id).First(&attachment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attachment, nil
}

func GetPlaygroundAttachmentForUser(id string, userID int) (*PlaygroundAttachment, error) {
	var attachment PlaygroundAttachment
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&attachment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attachment, nil
}

func ListPlaygroundAttachmentsForSession(userID int, sessionID string, now int64) ([]*PlaygroundAttachment, error) {
	var attachments []*PlaygroundAttachment
	err := DB.Where("user_id = ? AND session_id = ? AND status = ? AND expires_at > ?", userID, sessionID, PlaygroundAttachmentStatusActive, now).
		Order("created_at asc, id asc").
		Find(&attachments).Error
	return attachments, err
}

func CountActivePlaygroundAttachmentsForSession(userID int, sessionID string, now int64) (int64, error) {
	var count int64
	err := DB.Model(&PlaygroundAttachment{}).
		Where("user_id = ? AND session_id = ? AND status = ? AND expires_at > ?", userID, sessionID, PlaygroundAttachmentStatusActive, now).
		Count(&count).Error
	return count, err
}

func MarkPlaygroundAttachmentDeleted(id string, userID int, now int64) (bool, error) {
	result := DB.Model(&PlaygroundAttachment{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, PlaygroundAttachmentStatusActive).
		Updates(map[string]any{
			"status":       PlaygroundAttachmentStatusDeleted,
			"last_used_at": now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func MarkPlaygroundAttachmentExpired(id string, now int64) error {
	return DB.Model(&PlaygroundAttachment{}).
		Where("id = ? AND status = ?", id, PlaygroundAttachmentStatusActive).
		Updates(map[string]any{
			"status":       PlaygroundAttachmentStatusExpired,
			"last_used_at": now,
		}).Error
}

func TouchPlaygroundAttachmentLastUsed(id string, now int64) error {
	return DB.Model(&PlaygroundAttachment{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

func FindExpiredPlaygroundAttachments(now int64, limit int) ([]*PlaygroundAttachment, error) {
	if limit <= 0 {
		limit = 100
	}
	var attachments []*PlaygroundAttachment
	err := DB.Where("status = ? AND expires_at <= ?", PlaygroundAttachmentStatusActive, now).
		Order("expires_at asc, id asc").
		Limit(limit).
		Find(&attachments).Error
	return attachments, err
}

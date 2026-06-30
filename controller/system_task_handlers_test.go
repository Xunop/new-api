package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
)

func TestPlaygroundAttachmentCleanupHandlerRunsWhenUploadsDisabled(t *testing.T) {
	settings := system_setting.GetPlaygroundAttachmentSettings()
	original := *settings
	settings.Enabled = false
	t.Cleanup(func() {
		*settings = original
	})

	assert.True(t, playgroundAttachmentCleanupHandler{}.Enabled())
}

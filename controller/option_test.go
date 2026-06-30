package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type optionListResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"data"`
}

func TestGetOptionsHidesPlaygroundAttachmentSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	common.OptionMapRWMutex.Lock()
	original := common.OptionMap
	common.OptionMap = map[string]string{
		"playground_attachment.enabled":         "true",
		"playground_attachment.storage_driver":  "oss",
		"playground_attachment.oss_endpoint":    "oss-cn-hangzhou.aliyuncs.com",
		"playground_attachment.oss_bucket":      "test-bucket",
		"playground_attachment.oss_api_key":     "AKIA_TEST_VALUE",
		"playground_attachment.oss_secret":      "SECRET_TEST_VALUE",
		"playground_attachment.oss_object_prefix": "playground",
	}
	common.OptionMapRWMutex.Unlock()

	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = original
		common.OptionMapRWMutex.Unlock()
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/option/", nil)

	GetOptions(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload optionListResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)

	optionsByKey := make(map[string]string, len(payload.Data))
	for _, option := range payload.Data {
		optionsByKey[option.Key] = option.Value
	}

	assert.Equal(t, "true", optionsByKey["playground_attachment.enabled"])
	assert.Equal(t, "oss", optionsByKey["playground_attachment.storage_driver"])
	assert.Equal(
		t,
		"oss-cn-hangzhou.aliyuncs.com",
		optionsByKey["playground_attachment.oss_endpoint"],
	)
	assert.Equal(t, "test-bucket", optionsByKey["playground_attachment.oss_bucket"])
	assert.Equal(
		t,
		"playground",
		optionsByKey["playground_attachment.oss_object_prefix"],
	)
	_, hasAPIKey := optionsByKey["playground_attachment.oss_api_key"]
	_, hasSecret := optionsByKey["playground_attachment.oss_secret"]
	assert.False(t, hasAPIKey)
	assert.False(t, hasSecret)
}

package router

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRelayRouterRegistersPlaygroundAttachmentRoutesWithoutConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	require.NotPanics(t, func() {
		SetRelayRouter(engine)
	})
}

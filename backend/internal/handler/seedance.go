package handler

import (
	"mime"
	"net/http"
	"strings"

	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// SeedanceTasks exposes Ark's native asynchronous video task protocol.
func (h *OpenAIGatewayHandler) SeedanceTasks(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || !strings.EqualFold(mediaType, "application/json") {
			h.errorResponse(c, http.StatusUnsupportedMediaType, "invalid_request_error", "Seedance requires application/json")
			return
		}
	}
	key, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || key == nil || key.Group == nil {
		h.errorResponse(c, http.StatusForbidden, "permission_error", "Seedance requires an OpenAI or Custom group")
		return
	}
	targetPlatform := ""
	switch key.Group.Platform {
	case service.PlatformOpenAI:
		targetPlatform = service.PlatformOpenAI
	case service.PlatformComposite, service.PlatformCustom:
		targetPlatform = service.PlatformCustom
	default:
		h.errorResponse(c, http.StatusForbidden, "permission_error", "Seedance requires an OpenAI or Custom group")
		return
	}

	setActualUpstreamEndpoint(c, EndpointSeedanceTasks)
	endpoint := service.SeedanceEndpointCreate
	requestID := ""
	if c.Request.Method != http.MethodPost {
		taskID := strings.TrimSpace(c.Param("task_id"))
		if taskID == "" {
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
			return
		}
		requestID = service.SeedanceTaskKey(taskID)
		endpoint = service.SeedanceEndpointStatus
		if c.Request.Method == http.MethodDelete {
			endpoint = service.SeedanceEndpointDelete
		}
	}
	h.handleGrokMedia(c, endpoint, requestID, targetPlatform)
}

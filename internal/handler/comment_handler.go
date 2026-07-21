package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/middleware"
	"github.com/muunatic/schub/internal/service"
	"github.com/muunatic/schub/internal/utils"
)

type CommentHandler struct {
	platformService *service.PlatformService
	cfg             *config.Config
}

func NewCommentHandler(platformService *service.PlatformService, cfg *config.Config) *CommentHandler {
	return &CommentHandler{
		platformService: platformService,
		cfg:             cfg,
	}
}

func (h *CommentHandler) FetchComments(c *gin.Context) {
	platform := c.Param("platform")
	userID := c.GetString(middleware.ContextUserID)

	var comments interface{}
	var err error

	switch platform {
	case "youtube":
		comments, err = h.platformService.FetchYouTubeComments(c.Request.Context(), userID)
	case "tiktok":
		comments, err = h.platformService.FetchTikTokComments(c.Request.Context(), userID)
	case "twitter":
		comments, err = h.platformService.FetchTwitterComments(c.Request.Context(), userID)
	default:
		utils.BadRequest(c, "Bad Request")
		return
	}

	if err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessOKWithData(c, "OK", comments)
}

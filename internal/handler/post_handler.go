package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/middleware"
	"github.com/muunatic/schub/internal/service"
	"github.com/muunatic/schub/internal/utils"
)

type PostHandler struct {
	platformService *service.PlatformService
	cfg             *config.Config
}

func NewPostHandler(platformService *service.PlatformService, cfg *config.Config) *PostHandler {
	return &PostHandler{
		platformService: platformService,
		cfg:             cfg,
	}
}

func (h *PostHandler) FetchPosts(c *gin.Context) {
	platform := c.Param("platform")
	userID := c.GetString(middleware.ContextUserID)

	var posts interface{}
	var err error

	switch platform {
	case "youtube":
		posts, err = h.platformService.FetchYouTubePosts(c.Request.Context(), userID)
	case "tiktok":
		posts, err = h.platformService.FetchTikTokPosts(c.Request.Context(), userID)
	case "twitter":
		posts, err = h.platformService.FetchTwitterPosts(c.Request.Context(), userID)
	default:
		utils.BadRequest(c, "Bad Request")
		return
	}

	if err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessOKWithData(c, "OK", posts)
}

func (h *PostHandler) UploadVideo(c *gin.Context) {
	platform := c.Param("platform")
	userID := c.GetString(middleware.ContextUserID)

	title := c.PostForm("title")
	if title == "" {
		utils.BadRequest(c, "Bad Request")
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		utils.BadRequest(c, "Bad Request")
		return
	}
	defer file.Close()

	fileData := make([]byte, header.Size)
	if _, err := file.Read(fileData); err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	switch platform {
	case "youtube":
		_, err = h.platformService.UploadYouTubeVideo(c.Request.Context(), userID, title, fileData)
	case "tiktok":
		_, err = h.platformService.UploadTikTokVideo(c.Request.Context(), userID, title, fileData, header.Size)
	case "twitter":
		fileMime := header.Header.Get("Content-Type")
		if fileMime == "" {
			fileMime = "video/mp4"
		}
		_, err = h.platformService.UploadTwitterVideo(c.Request.Context(), userID, title, fileData, fileMime)
	default:
		utils.BadRequest(c, "Bad Request")
		return
	}

	if err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessCreated(c, "Created")
}

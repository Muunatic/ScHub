package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/middleware"
	"github.com/muunatic/schub/internal/service"
	"github.com/muunatic/schub/internal/utils"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
	cfg          *config.Config
}

func NewOAuthHandler(oauthService *service.OAuthService, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		cfg:          cfg,
	}
}

func (h *OAuthHandler) TikTokAuth(c *gin.Context) {
	url := h.oauthService.GetTikTokAuthURL()
	c.Redirect(302, url)
}

func (h *OAuthHandler) TikTokCallback(c *gin.Context) {
	code := c.Query("code")
	userID := c.GetString(middleware.ContextUserID)

	if err := h.oauthService.HandleTikTokCallback(c.Request.Context(), code, userID); err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessCreatedWithRedirect(c, "Created", c.GetHeader("Origin")+"/dashboard")
}

func (h *OAuthHandler) TwitterAuth(c *gin.Context) {
	url := h.oauthService.GetTwitterAuthURL()
	c.Redirect(302, url)
}

func (h *OAuthHandler) TwitterCallback(c *gin.Context) {
	code := c.Query("code")
	userID := c.GetString(middleware.ContextUserID)

	if err := h.oauthService.HandleTwitterCallback(c.Request.Context(), code, userID); err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessCreatedWithRedirect(c, "Created", c.GetHeader("Origin")+"/dashboard")
}

func (h *OAuthHandler) YouTubeAuth(c *gin.Context) {
	url := h.oauthService.GetYouTubeAuthURL()
	c.Redirect(302, url)
}

func (h *OAuthHandler) YouTubeCallback(c *gin.Context) {
	code := c.Query("code")
	userID := c.GetString(middleware.ContextUserID)

	if err := h.oauthService.HandleYouTubeCallback(c.Request.Context(), code, userID); err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	utils.SuccessCreatedWithRedirect(c, "Created", c.GetHeader("Origin")+"/dashboard")
}

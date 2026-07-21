package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/middleware"
	"github.com/muunatic/schub/internal/repository"
)

type ViewHandler struct {
	userRepo    *repository.UserRepository
	accountRepo *repository.AccountRepository
	apiKeyRepo  *repository.APIKeyRepository
	cfg         *config.Config
}

func NewViewHandler(
	userRepo *repository.UserRepository,
	accountRepo *repository.AccountRepository,
	apiKeyRepo *repository.APIKeyRepository,
	cfg *config.Config,
) *ViewHandler {
	return &ViewHandler{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		apiKeyRepo:  apiKeyRepo,
		cfg:         cfg,
	}
}

func (h *ViewHandler) HeadView(c *gin.Context) {
	_, err := h.userRepo.FindAllUsernames(c.Request.Context())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ViewHandler) HomeView(c *gin.Context) {
	c.HTML(http.StatusOK, "home/index.html", nil)
}

func (h *ViewHandler) DashboardView(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)

	user, err := h.userRepo.SelectID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	accounts, err := h.accountRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	platforms := make([]string, 0)
	for _, acc := range accounts {
		platforms = append(platforms, string(acc.Platform))
	}

	c.HTML(http.StatusOK, "dashboard/index.html", gin.H{
		"user":     user,
		"platform": platforms,
	})
}

func (h *ViewHandler) AccountView(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)

	user, err := h.userRepo.SelectID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	accounts, err := h.accountRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	apiKey, err := h.apiKeyRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("http://%s", c.Request.Host)
	if h.cfg.Env == "PRODUCTION" || strings.Contains(c.Request.Host, ".ngrok-free.app") {
		url = fmt.Sprintf("https://%s", c.Request.Host)
	}

	c.HTML(http.StatusOK, "account/index.html", gin.H{
		"user":    user,
		"account": accounts,
		"apiKey":  apiKey,
		"url":     url,
	})
}

func (h *ViewHandler) PostView(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)

	sortBy := c.Query("sortby")
	sort := c.Query("sort")
	platform := c.DefaultQuery("platform", "tiktok,twitter,youtube")

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	var baseURL string
	if sortBy != "" && sort != "" {
		baseURL = fmt.Sprintf("%s://%s/api/v1/post?platform=%s&sort=%s&sortby=%s", scheme, c.Request.Host, platform, sort, sortBy)
	} else {
		baseURL = fmt.Sprintf("%s://%s/api/v1/post?platform=%s", scheme, c.Request.Host, platform)
	}

	req, _ := http.NewRequest("GET", baseURL, nil)
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var posts interface{}
	if err := json.Unmarshal(body, &posts); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	accounts, err := h.accountRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	platforms := make([]string, 0)
	for _, acc := range accounts {
		platforms = append(platforms, string(acc.Platform))
	}

	fmt.Printf("Data posts: %#v\n", posts)

	c.HTML(http.StatusOK, "post/index.html", gin.H{
		"data":     posts,
		"platform": platforms,
	})
}

func (h *ViewHandler) CommentView(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)

	sortBy := c.Query("sortby")
	sort := c.Query("sort")
	platform := c.DefaultQuery("platform", "tiktok,twitter,youtube")

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	var baseURL string
	if sortBy != "" && sort != "" {
		baseURL = fmt.Sprintf("%s://%s/api/v1/comment?platform=%s&sort=%s&sortby=%s", scheme, c.Request.Host, platform, sort, sortBy)
	} else {
		baseURL = fmt.Sprintf("%s://%s/api/v1/comment?platform=%s", scheme, c.Request.Host, platform)
	}

	req, _ := http.NewRequest("GET", baseURL, nil)
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var comments interface{}
	if err := json.Unmarshal(body, &comments); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	accounts, err := h.accountRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	platforms := make([]string, 0)
	for _, acc := range accounts {
		platforms = append(platforms, string(acc.Platform))
	}

	c.HTML(http.StatusOK, "comment/index.html", gin.H{
		"data":     comments,
		"platform": platforms,
	})
}

// func (h *ViewHandler) VerifyView(c *gin.Context) {
// 	c.HTML(http.StatusOK, "verify/index.html", nil)
// }

func (h *ViewHandler) LoginView(c *gin.Context) {
	url := fmt.Sprintf("http://%s", c.Request.Host)
	if h.cfg.Env == "PRODUCTION" || strings.Contains(c.Request.Host, ".ngrok-free.app") {
		url = fmt.Sprintf("https://%s", c.Request.Host)
	}

	c.HTML(http.StatusOK, "login/index.html", gin.H{
		"url": url,
	})
}

func (h *ViewHandler) RegisterView(c *gin.Context) {
	url := fmt.Sprintf("http://%s", c.Request.Host)
	if h.cfg.Env == "PRODUCTION" || strings.Contains(c.Request.Host, ".ngrok-free.app") {
		url = fmt.Sprintf("https://%s", c.Request.Host)
	}

	c.HTML(http.StatusOK, "register/index.html", gin.H{
		"url": url,
	})
}

func (h *ViewHandler) PrivacyView(c *gin.Context) {
	c.HTML(http.StatusOK, "privacy/index.html", nil)
}

func (h *ViewHandler) TermsView(c *gin.Context) {
	c.HTML(http.StatusOK, "terms/index.html", nil)
}

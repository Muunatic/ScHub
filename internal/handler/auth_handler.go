package handler

import (
	// "encoding/json"
	"net/http"
	// "strings"

	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/service"
	"github.com/muunatic/schub/internal/utils"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	// utils.Locked(c, "Closed")
	// return

	var req struct {
		Email    string `json:"email" form:"email"`
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBind(&req); err != nil || req.Email == "" || req.Username == "" || req.Password == "" {
		utils.BadRequest(c, "Bad Request")
		return
	}
	if _, err := h.authService.Register(c.Request.Context(), req.Email, req.Username, req.Password); err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.SuccessCreatedWithRedirect(c, "Created", c.GetHeader("Origin")+"/login")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBind(&req); err != nil || req.Email == "" || req.Password == "" {
		utils.BadRequest(c, "Bad Request")
		return
	}

	token, _, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	domain := "localhost"
	if h.cfg.Env == "PRODUCTION" {
		domain = h.cfg.Domain
	}

	c.SetCookie("token", token, 0, "/", domain, true, true)
	utils.SuccessOKWithRedirect(c, "Cookie is set.", c.GetHeader("Origin")+"/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	domain := "localhost"
	if h.cfg.Env == "PRODUCTION" {
		domain = h.cfg.Domain
	}

	c.SetCookie("token", "", -1, "/", domain, true, true)
	c.JSON(http.StatusOK, gin.H{
		"status":      200,
		"message":     "User cookie has been deleted.",
		"redirectUrl": c.GetHeader("Origin"),
	})
}

// Archived
// func (h *AuthHandler) Verify(c *gin.Context) {
// 	var req struct {
// 		Password string `json:"password" form:"password"`
// 	}
// 	if err := c.ShouldBind(&req); err != nil || req.Password == "" {
// 		utils.BadRequest(c, "Bad Request")
// 		return
// 	}

// 	reqBody := `{"password":"` + req.Password + `"}`
// 	verifyReq, err := http.NewRequest("POST", "https://schub.typeslint.com/v1/service/verify", strings.NewReader(reqBody))
// 	if err != nil {
// 		utils.InternalServerError(c, "Internal Server Error")
// 		return
// 	}
// 	verifyReq.Header.Set("Content-Type", "application/json")
// 	verifyReq.Header.Set("X-Api-Key", "X-API-Key")

// 	resp, err := http.DefaultClient.Do(verifyReq)
// 	if err != nil {
// 		utils.InternalServerError(c, "Internal Server Error")
// 		return
// 	}
// 	defer resp.Body.Close()

// 	var data struct {
// 		Token   string `json:"token"`
// 		Message string `json:"message"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
// 		utils.InternalServerError(c, "Internal Server Error")
// 		return
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		msg := data.Message
// 		if msg == "" {
// 			msg = "Invalid password!"
// 		}
// 		utils.Unauthorized(c, msg)
// 		return
// 	}

// 	c.SetCookie("entry_token", data.Token, 0, "/", c.Request.Host, false, true)
// 	utils.SuccessOK(c, "OK")
// }

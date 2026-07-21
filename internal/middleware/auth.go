package middleware

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/models"
	"github.com/muunatic/schub/internal/repository"
	"github.com/muunatic/schub/internal/utils"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	userRepo   *repository.UserRepository
	apiKeyRepo *repository.APIKeyRepository
	db         *gorm.DB
	cfg        *config.Config
}

func NewAuthMiddleware(db *gorm.DB, userRepo *repository.UserRepository, apiKeyRepo *repository.APIKeyRepository, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
		db:         db,
		cfg:        cfg,
	}
}

func (m *AuthMiddleware) Restrict() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("token")
		authorization := c.GetHeader("Authorization")
		referer := c.GetHeader("Referer")

		if referer != "" {
			if strings.HasPrefix(referer, m.cfg.ProdURL+"/oas") ||
				strings.HasPrefix(referer, m.cfg.ProdURL+"/oas") {
				m.authenticateWithAPIKey(c, authorization)
				return
			}
		}

		if (token == "" && authorization == "") || (token != "" && authorization != "") {
			utils.Unauthorized(c, "Token not provided!")
			c.Abort()
			return
		}

		if token != "" {
			m.authenticateWithJWT(c, token)
		} else {
			m.authenticateWithAPIKey(c, authorization)
		}
	}
}

func (m *AuthMiddleware) authenticateWithJWT(c *gin.Context, tokenString string) {
	jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.cfg.JWTSecret), nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "signature") {
			domain := "localhost"
			if m.cfg.Env == "PRODUCTION" {
				domain = m.cfg.Domain
			}
			c.SetCookie("token", "", -1, "/", domain, true, true)
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
			c.Redirect(302, fmt.Sprintf("%s://%s/login", scheme, c.Request.Host))
			c.Abort()
			return
		}
		utils.Unauthorized(c, err.Error())
		c.Abort()
		return
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		utils.Unauthorized(c, "Invalid token")
		c.Abort()
		return
	}

	id, ok := claims["id"].(string)
	if !ok || id == "" {
		utils.Unauthorized(c, "Invalid token payload")
		c.Abort()
		return
	}

	user, err := m.userRepo.SelectID(c.Request.Context(), id)
	if err != nil {
		utils.NotFound(c, "Not Found")
		c.Abort()
		return
	}

	c.Set("userID", user.ID)
	c.Next()
}

func (m *AuthMiddleware) authenticateWithAPIKey(c *gin.Context, authorization string) {
	if !strings.HasPrefix(authorization, "Bearer ") {
		utils.BadRequest(c, "Authorization header must start with 'Bearer'")
		c.Abort()
		return
	}

	key := strings.TrimPrefix(authorization, "Bearer ")
	apiKey, err := m.apiKeyRepo.FindByKey(c.Request.Context(), key)
	if err != nil {
		utils.BadRequest(c, "Invalid Key")
		c.Abort()
		return
	}

	user, err := m.userRepo.SelectID(c.Request.Context(), apiKey.UserID)
	if err != nil {
		utils.NotFound(c, "Not Found")
		c.Abort()
		return
	}

	c.Set("userID", user.ID)
	c.Next()
}

func (m *AuthMiddleware) VerifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		entryToken, _ := c.Cookie("entry_token")

		fmt.Println(c.Request.Method, c.Request.URL.Path)

		if entryToken != "" {
			jwtToken, err := jwt.Parse(entryToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(m.cfg.APISecret), nil
			})

			if err != nil {
				utils.Unauthorized(c, err.Error())
				c.Abort()
				return
			}

			if jwtToken.Valid {
				c.Next()
				return
			}
		}

		if entryToken == "" {
			if c.Request.URL.Path == "/v1/verify" {
				c.Next()
				return
			}
			if c.Request.URL.Path == "/verify" {
				c.HTML(200, "verify/index.html", nil)
				c.Abort()
				return
			}
			c.Redirect(302, "/verify")
			c.Abort()
			return
		}

		c.Next()
	}
}

func FilterQueryParams() gin.HandlerFunc {
	validPlatforms := map[string]bool{
		"tiktok":  true,
		"twitter": true,
		"youtube": true,
	}

	return func(c *gin.Context) {
		rawPlatform := c.DefaultQuery("platform", "tiktok,twitter,youtube")

		if rawPlatform == "" {
			c.Next()
			return
		}

		platforms := strings.Split(rawPlatform, ",")

		for _, p := range platforms {
			p = strings.TrimSpace(strings.ToLower(p))
			if !validPlatforms[p] {
				utils.BadRequest(c, "Invalid platform, valid options are: tiktok, twitter, youtube")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
		if len(parts) < 4 {
			c.Next()
			return
		}

		rawPlatform := strings.ToLower(parts[2])
		platformMap := map[string]string{
			"tiktok":  "TikTok",
			"twitter": "Twitter",
			"youtube": "YouTube",
		}

		platform, ok := platformMap[rawPlatform]
		if !ok {
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		var account models.Account
		err := m.db.WithContext(ctx).
			Where("\"userId\" = ? AND platform = ?", uid, platform).
			First(&account).Error

		if err != nil {
			utils.NotFound(c, "Not Found")
			c.Abort()
			return
		}

		switch platform {
		case "YouTube":
			m.refreshYouTubeToken(c, uid, *account.RefreshToken)
		case "TikTok":
			m.refreshTikTokToken(c, uid, *account.RefreshToken)
		case "Twitter":
			m.refreshTwitterToken(c, uid, *account.RefreshToken)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) refreshYouTubeToken(c *gin.Context, userID, refreshToken string) {
	form := url.Values{}
	form.Set("client_id", m.cfg.GoogleClientID)
	form.Set("client_secret", m.cfg.GoogleClientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	resp, err := doHTTPPost(c, config.YouTubeTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	if !resp.OK {
		m.db.WithContext(c.Request.Context()).
			Where("\"userId\" = ? AND platform = ?", userID, "YouTube").
			Delete(&models.Account{})
		return
	}

	tokenData := resp.Data
	at, _ := tokenData["access_token"].(string)
	rt, _ := tokenData["refresh_token"].(string)
	if rt == "" {
		rt = refreshToken
	}

	m.db.WithContext(c.Request.Context()).
		Model(&models.Account{}).
		Where("\"userId\" = ? AND platform = ?", userID, "YouTube").
		Updates(map[string]interface{}{
			"accessToken":  at,
			"refreshToken": rt,
		})
}

func (m *AuthMiddleware) refreshTikTokToken(c *gin.Context, userID, refreshToken string) {
	form := url.Values{}
	form.Set("client_key", m.cfg.TikTokClientID)
	form.Set("client_secret", m.cfg.TikTokClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	resp, err := doHTTPPost(c, config.TikTokTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	if !resp.OK {
		m.db.WithContext(c.Request.Context()).
			Where("\"userId\" = ? AND platform = ?", userID, "TikTok").
			Delete(&models.Account{})
		return
	}

	tokenData := resp.Data
	at, _ := tokenData["access_token"].(string)
	rt, _ := tokenData["refresh_token"].(string)
	if rt == "" {
		rt = refreshToken
	}

	m.db.WithContext(c.Request.Context()).
		Model(&models.Account{}).
		Where("\"userId\" = ? AND platform = ?", userID, "TikTok").
		Updates(map[string]interface{}{
			"accessToken":  at,
			"refreshToken": rt,
		})
}

func (m *AuthMiddleware) refreshTwitterToken(c *gin.Context, userID, refreshToken string) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	auth := base64.StdEncoding.EncodeToString([]byte(m.cfg.TwitterClientID + ":" + m.cfg.TwitterClientSecret))

	resp, err := doHTTPPostWithAuth(c, config.TwitterTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()), "Basic "+auth)
	if err != nil {
		return
	}

	if !resp.OK {
		m.db.WithContext(c.Request.Context()).
			Where("\"userId\" = ? AND platform = ?", userID, "Twitter").
			Delete(&models.Account{})
		return
	}

	tokenData := resp.Data
	at, _ := tokenData["access_token"].(string)
	rt, _ := tokenData["refresh_token"].(string)
	if rt == "" {
		rt = refreshToken
	}

	m.db.WithContext(c.Request.Context()).
		Model(&models.Account{}).
		Where("\"userId\" = ? AND platform = ?", userID, "Twitter").
		Updates(map[string]interface{}{
			"accessToken":  at,
			"refreshToken": rt,
		})
}

const ContextUserID = "userID"

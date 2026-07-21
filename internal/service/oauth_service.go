package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/dto"
	"github.com/muunatic/schub/internal/models"
	"github.com/muunatic/schub/internal/repository"
)

type OAuthService struct {
	accountRepo *repository.AccountRepository
	cfg         *config.Config
}

func NewOAuthService(accountRepo *repository.AccountRepository, cfg *config.Config) *OAuthService {
	return &OAuthService{
		accountRepo: accountRepo,
		cfg:         cfg,
	}
}

func (s *OAuthService) GetTikTokAuthURL() string {
	return fmt.Sprintf("%s?client_key=%s&redirect_uri=%s/%s&scope=user.info.basic,user.info.profile,user.info.stats,video.list,video.upload,video.publish&response_type=code",
		config.TikTokAuthURL, s.cfg.TikTokClientID, s.cfg.ProdURL, s.cfg.TikTokRedirectURI)
}

func (s *OAuthService) HandleTikTokCallback(ctx context.Context, code, userID string) error {
	tokenData, err := s.exchangeTikTokCode(code)
	if err != nil {
		return err
	}

	accessToken, _ := tokenData["access_token"].(string)
	refreshToken, _ := tokenData["refresh_token"].(string)

	userInfo, err := s.getTikTokUserInfo(accessToken)
	if err != nil {
		return err
	}

	username := userInfo.User.Username
	acc := &models.Account{
		Username:     &username,
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		Platform:     models.PlatformTikTok,
		UserID:       userID,
	}

	return s.accountRepo.Upsert(ctx, acc)
}

func (s *OAuthService) exchangeTikTokCode(code string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("client_key", s.cfg.TikTokClientID)
	form.Set("client_secret", s.cfg.TikTokClientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", fmt.Sprintf("%s/%s", s.cfg.ProdURL, s.cfg.TikTokRedirectURI))

	return s.doFormPost(config.TikTokTokenURL, form)
}

func (s *OAuthService) getTikTokVideos(accessToken string) ([]dto.TikTokVideo, error) {
	body := strings.NewReader(`{"max_count": 10}`)
	req, err := http.NewRequest("POST", "https://open.tiktokapis.com/v2/video/list/?fields=id,title,cover_image_url,like_count,comment_count,share_count", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.TikTokVideoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data.Videos, nil
}

func (s *OAuthService) getTikTokUserInfo(accessToken string) (*dto.TikTokUserInfoData, error) {
	req, err := http.NewRequest("GET", "https://open.tiktokapis.com/v2/user/info/?fields=display_name,username,avatar_url", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.TikTokUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *OAuthService) GetTwitterAuthURL() string {
	return fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=tweet.read+tweet.write+users.email+users.read+follows.read+follows.write+offline.access&state=random_state&code_challenge=challenge&code_challenge_method=plain",
		config.TwitterAuthURL, s.cfg.TwitterClientID, url.QueryEscape(s.cfg.ProdURL+"/"+s.cfg.TwitterRedirectURI))
}

func (s *OAuthService) HandleTwitterCallback(ctx context.Context, code, userID string) error {
	tokenData, err := s.exchangeTwitterCode(code)
	if err != nil {
		return err
	}

	accessToken, _ := tokenData["access_token"].(string)
	refreshToken, _ := tokenData["refresh_token"].(string)

	profile, err := s.getTwitterProfile(accessToken)
	if err != nil {
		return err
	}

	username := profile.Data.ID
	email := profile.Data.ConfirmedEmail

	acc := &models.Account{
		Username:     &username,
		Email:        &email,
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		Platform:     models.PlatformTwitter,
		UserID:       userID,
	}

	return s.accountRepo.Upsert(ctx, acc)
}

func (s *OAuthService) exchangeTwitterCode(code string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", s.cfg.TwitterClientID)
	form.Set("client_secret", s.cfg.TwitterClientSecret)
	form.Set("redirect_uri", s.cfg.ProdURL+"/"+s.cfg.TwitterRedirectURI)
	form.Set("code_verifier", "challenge")
	form.Set("scope", "tweet.read tweet.write users.email users.read follows.read follows.write offline.access")

	auth := base64.StdEncoding.EncodeToString([]byte(s.cfg.TwitterClientID + ":" + s.cfg.TwitterClientSecret))

	req, err := http.NewRequest("POST", config.TwitterTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if errMsg, ok := result["error"]; ok {
		return nil, fmt.Errorf("twitter token error: %v", errMsg)
	}

	return result, nil
}

func (s *OAuthService) getTwitterProfile(accessToken string) (*TwitterProfileResponse, error) {
	req, err := http.NewRequest("GET", "https://api.twitter.com/2/users/me?user.fields=confirmed_email", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TwitterProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *OAuthService) GetTwitterOAuth1RequestToken() (string, error) {
	oauth := NewOAuth1(s.cfg.TwitterAppKey, s.cfg.TwitterAppSecretKey)
	callbackURL := s.cfg.ProdURL + "/" + s.cfg.TwitterRedirectURIOAuth1

	requestData := map[string]string{
		"oauth_callback": callbackURL,
	}

	header := oauth.Authorize("https://api.twitter.com/oauth/request_token", "POST", requestData)

	form := url.Values{}
	form.Set("oauth_callback", callbackURL)

	req, err := http.NewRequest("POST", "https://api.twitter.com/oauth/request_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	params, _ := url.ParseQuery(string(body))
	oauthToken := params.Get("oauth_token")

	return oauthToken, nil
}

func (s *OAuthService) HandleTwitterOAuth1Callback(ctx context.Context, oauthToken, oauthVerifier, userID string) error {
	oauth := NewOAuth1(s.cfg.TwitterAppKey, s.cfg.TwitterAppSecretKey)

	requestData := map[string]string{
		"oauth_token":    oauthToken,
		"oauth_verifier": oauthVerifier,
	}

	header := oauth.Authorize("https://api.twitter.com/oauth/access_token", "POST", requestData)

	form := url.Values{}
	form.Set("oauth_token", oauthToken)
	form.Set("oauth_verifier", oauthVerifier)

	req, err := http.NewRequest("POST", "https://api.twitter.com/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	params, _ := url.ParseQuery(string(body))
	accessToken := params.Get("oauth_token")
	accessTokenSecret := params.Get("oauth_token_secret")

	acc := &models.Account{
		AccessKey:       &accessToken,
		AccessSecretKey: &accessTokenSecret,
		Platform:        models.PlatformTwitter,
		UserID:          userID,
	}

	return s.accountRepo.Upsert(ctx, acc)
}

func (s *OAuthService) GetYouTubeAuthURL() string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s/%s&response_type=code&scope=https://www.googleapis.com/auth/youtube.readonly+https://www.googleapis.com/auth/userinfo.email+https://www.googleapis.com/auth/youtube.force-ssl+https://www.googleapis.com/auth/youtube.upload&access_type=offline&prompt=consent",
		config.YouTubeAuthURL, s.cfg.GoogleClientID, s.cfg.ProdURL, s.cfg.GoogleRedirectURI)
}

func (s *OAuthService) HandleYouTubeCallback(ctx context.Context, code, userID string) error {
	tokenData, err := s.exchangeYouTubeCode(code)
	if err != nil {
		return err
	}

	accessToken, _ := tokenData["access_token"].(string)
	refreshToken, _ := tokenData["refresh_token"].(string)

	channelID, err := s.getYouTubeChannelID(accessToken)
	if err != nil {
		return err
	}

	email, err := s.getYouTubeProfileEmail(accessToken)
	if err != nil {
		return err
	}

	acc := &models.Account{
		Username:     &channelID,
		Email:        &email,
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		Platform:     models.PlatformYouTube,
		UserID:       userID,
	}

	return s.accountRepo.Upsert(ctx, acc)
}

func (s *OAuthService) exchangeYouTubeCode(code string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.cfg.GoogleClientID)
	form.Set("client_secret", s.cfg.GoogleClientSecret)
	form.Set("redirect_uri", s.cfg.ProdURL+"/"+s.cfg.GoogleRedirectURI)
	form.Set("grant_type", "authorization_code")

	return s.doFormPost(config.YouTubeTokenURL, form)
}

func (s *OAuthService) getYouTubeChannelID(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/channels?part=snippet&mine=true", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var channelResp dto.YouTubeChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&channelResp); err != nil {
		return "", err
	}

	if len(channelResp.Items) == 0 {
		return "", fmt.Errorf("no channel found")
	}

	return channelResp.Items[0].ID, nil
}

func (s *OAuthService) getYouTubeProfileEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var profile dto.YouTubeProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}

	return profile.Email, nil
}

func (s *OAuthService) doFormPost(urlStr string, form url.Values) (map[string]interface{}, error) {
	resp, err := http.PostForm(urlStr, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

type TwitterProfileResponse struct {
	Data TwitterProfileData `json:"data"`
}

type TwitterProfileData struct {
	ID             string `json:"id"`
	ConfirmedEmail string `json:"confirmed_email"`
}
type OAuth1 struct {
	consumerKey    string
	consumerSecret string
}

func NewOAuth1(consumerKey, consumerSecret string) *OAuth1 {
	return &OAuth1{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
	}
}

func (o *OAuth1) Authorize(requestURL, method string, data map[string]string) map[string]string {
	nonce := generateNonce()
	timestamp := generateTimestamp()

	params := url.Values{}
	params.Set("oauth_consumer_key", o.consumerKey)
	params.Set("oauth_nonce", nonce)
	params.Set("oauth_signature_method", "HMAC-SHA1")
	params.Set("oauth_timestamp", timestamp)
	params.Set("oauth_version", "1.0")

	for k, v := range data {
		params.Set(k, v)
	}

	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(requestURL) + "&" + url.QueryEscape(params.Encode())

	mac := hmac.New(sha1.New, []byte(o.consumerSecret+"&"))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params.Set("oauth_signature", signature)

	var parts []string
	for k := range params {
		if strings.HasPrefix(k, "oauth_") {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, k, url.QueryEscape(params.Get(k))))
		}
	}

	return map[string]string{
		"Authorization": "OAuth " + strings.Join(parts, ", "),
	}
}

func (o *OAuth1) AuthorizeWithToken(requestURL, method string, data map[string]string, tokenKey, tokenSecret string) map[string]string {
	nonce := generateNonce()
	timestamp := generateTimestamp()

	params := url.Values{}
	params.Set("oauth_consumer_key", o.consumerKey)
	params.Set("oauth_nonce", nonce)
	params.Set("oauth_signature_method", "HMAC-SHA1")
	params.Set("oauth_timestamp", timestamp)
	params.Set("oauth_token", tokenKey)
	params.Set("oauth_version", "1.0")

	for k, v := range data {
		params.Set(k, v)
	}

	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(requestURL) + "&" + url.QueryEscape(params.Encode())

	mac := hmac.New(sha1.New, []byte(o.consumerSecret+"&"+tokenSecret))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	params.Set("oauth_signature", signature)

	var parts []string
	for k := range params {
		if strings.HasPrefix(k, "oauth_") {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, k, url.QueryEscape(params.Get(k))))
		}
	}

	return map[string]string{
		"Authorization": "OAuth " + strings.Join(parts, ", "),
	}
}

func generateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func generateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

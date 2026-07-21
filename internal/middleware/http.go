package middleware

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpResult struct {
	OK   bool
	Data map[string]interface{}
}

func doHTTPPost(c *gin.Context, url, contentType string, body io.Reader) (*httpResult, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return &httpResult{OK: false}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, err
	}

	return &httpResult{OK: true, Data: data}, nil
}

func doHTTPPostWithAuth(c *gin.Context, url, contentType string, body io.Reader, authHeader string) (*httpResult, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return &httpResult{OK: false}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, err
	}

	return &httpResult{OK: true, Data: data}, nil
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Powered-By", "")
		c.Next()
	}
}

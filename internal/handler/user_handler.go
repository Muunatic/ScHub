package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/dto"
	"github.com/muunatic/schub/internal/utils"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) UploadPost(c *gin.Context) {
	title := c.PostForm("title")
	file, header, err := c.Request.FormFile("file")
	if err != nil || title == "" {
		utils.BadRequest(c, "Bad Request")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		utils.InternalServerError(c, "Internal Server Error")
		return
	}

	rawPlatform := c.DefaultQuery("platform", "")
	var platforms []string

	if rawPlatform != "" {
		platforms = strings.Split(rawPlatform, ",")
	} else {
		utils.BadRequest(c, "Platform query parameter required")
		return
	}

	type platformResult struct {
		Platform string      `json:"platform"`
		Status   int         `json:"status"`
		Message  string      `json:"message"`
		Data     interface{} `json:"data,omitempty"`
	}

	results := make([]platformResult, 0)

	for _, p := range platforms {
		p = strings.TrimSpace(strings.ToLower(p))
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s/api/v1/%s/post", scheme, c.Request.Host, p)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		w.WriteField("title", title)

		part, _ := w.CreateFormFile("file", header.Filename)
		part.Write(fileData)
		w.Close()

		req, _ := http.NewRequest("POST", baseURL, &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())

		for _, cookie := range c.Request.Cookies() {
			req.AddCookie(cookie)
		}

		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			results = append(results, platformResult{
				Platform: p,
				Status:   500,
				Message:  "Internal Server Error",
			})
			continue
		}
		defer resp.Body.Close()

		var resultData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&resultData)

		status := 500
		message := "Internal Server Error"
		if resultData != nil {
			if s, ok := resultData["status"].(float64); ok {
				status = int(s)
			}
			if m, ok := resultData["message"].(string); ok {
				message = m
			}
		}

		results = append(results, platformResult{
			Platform: p,
			Status:   status,
			Message:  message,
			Data:     resultData,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  201,
		"message": "Created",
		"data":    results,
	})
}

func (h *UserHandler) ListPost(c *gin.Context) {
	rawPlatform := c.DefaultQuery("platform", "tiktok")
	sortBy := c.Query("sortby")
	sortOrder := c.Query("sort")

	if sortBy != "" && sortOrder == "" {
		utils.BadRequest(c, "Parameter 'sort' harus disertakan jika 'sortBy' digunakan.")
		return
	}

	platforms := strings.Split(rawPlatform, ",")

	type postResponse struct {
		Status   int            `json:"status"`
		Message  string         `json:"message"`
		Platform string         `json:"platform"`
		Data     []dto.PostData `json:"data"`
	}

	results := make([]postResponse, 0)

	for _, p := range platforms {
		p = strings.TrimSpace(strings.ToLower(p))

		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s/api/v1/%s/post", scheme, c.Request.Host, p)

		req, _ := http.NewRequest("GET", baseURL, nil)

		for _, cookie := range c.Request.Cookies() {
			req.AddCookie(cookie)
		}

		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var data postResponse
		json.NewDecoder(resp.Body).Decode(&data)
		data.Platform = p
		results = append(results, data)
	}

	type mergedPost struct {
		dto.PostData
		Platform string `json:"platform"`
	}

	mergedData := make([]mergedPost, 0)
	for _, r := range results {
		for _, post := range r.Data {
			mergedData = append(mergedData, mergedPost{
				PostData: post,
				Platform: r.Platform,
			})
		}
	}

	if sortBy != "" && sortOrder != "" {
		sort.Slice(mergedData, func(i, j int) bool {
			a := mergedData[i]
			b := mergedData[j]

			switch sortBy {
			case "createdAt":
				ti, _ := time.Parse(time.RFC3339, a.CreatedAt)
				tj, _ := time.Parse(time.RFC3339, b.CreatedAt)
				if sortOrder == "asc" {
					return ti.Before(tj)
				}
				return ti.After(tj)
			case "comments":
				if sortOrder == "asc" {
					return a.Comments < b.Comments
				}
				return a.Comments > b.Comments
			case "likes":
				if sortOrder == "asc" {
					return a.Likes < b.Likes
				}
				return a.Likes > b.Likes
			default:
				return false
			}
		})
	}

	utils.SuccessOKWithData(c, "OK", mergedData)
}

func (h *UserHandler) ListComment(c *gin.Context) {
	rawPlatform := c.DefaultQuery("platform", "tiktok,twitter,youtube")
	sortBy := c.Query("sortby")
	sortOrder := c.Query("sort")

	if sortBy != "" && sortOrder == "" {
		utils.BadRequest(c, "Parameter 'sort' harus disertakan jika 'sortBy' digunakan.")
		return
	}

	platforms := strings.Split(rawPlatform, ",")

	type commentResponse struct {
		Status   int               `json:"status"`
		Message  string            `json:"message"`
		Platform string            `json:"platform"`
		Data     []dto.CommentData `json:"data"`
	}

	results := make([]commentResponse, 0)

	for _, p := range platforms {
		p = strings.TrimSpace(strings.ToLower(p))

		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s/api/v1/%s/comment", scheme, c.Request.Host, p)

		req, _ := http.NewRequest("GET", baseURL, nil)

		for _, cookie := range c.Request.Cookies() {
			req.AddCookie(cookie)
		}

		if auth := c.GetHeader("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var data commentResponse
		json.NewDecoder(resp.Body).Decode(&data)
		data.Platform = p
		results = append(results, data)
	}

	type mergedComment struct {
		dto.CommentData
		Platform string `json:"platform"`
	}

	mergedData := make([]mergedComment, 0)
	for _, r := range results {
		for _, comment := range r.Data {
			mergedData = append(mergedData, mergedComment{
				CommentData: comment,
				Platform:    r.Platform,
			})
		}
	}

	if sortBy != "" && sortOrder != "" {
		sort.Slice(mergedData, func(i, j int) bool {
			a := mergedData[i]
			b := mergedData[j]

			switch sortBy {
			case "createdAt":
				ti, _ := time.Parse(time.RFC3339, a.CreatedAt)
				tj, _ := time.Parse(time.RFC3339, b.CreatedAt)
				if sortOrder == "asc" {
					return ti.Before(tj)
				}
				return ti.After(tj)
			default:
				return false
			}
		})
	}

	utils.SuccessOKWithData(c, "OK", mergedData)
}

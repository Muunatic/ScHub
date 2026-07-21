package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/dto"
)

func SuccessOK(c *gin.Context, message string) {
	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  http.StatusOK,
		Message: message,
	})
}

func SuccessOKWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func SuccessCreated(c *gin.Context, message string) {
	c.JSON(http.StatusCreated, dto.APIResponse{
		Status:  http.StatusCreated,
		Message: message,
	})
}

func SuccessCreatedWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, dto.APIResponse{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

func SuccessCreatedWithRedirect(c *gin.Context, message, redirectURL string) {
	c.JSON(http.StatusCreated, dto.APIResponseWithRedirect{
		Status:      http.StatusCreated,
		Message:     message,
		RedirectURL: redirectURL,
	})
}

func SuccessOKWithRedirect(c *gin.Context, message, redirectURL string) {
	c.JSON(http.StatusOK, dto.APIResponseWithRedirect{
		Status:      http.StatusOK,
		Message:     message,
		RedirectURL: redirectURL,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, dto.APIResponse{
		Status:  http.StatusBadRequest,
		Message: message,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, dto.APIResponse{
		Status:  http.StatusUnauthorized,
		Message: message,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, dto.APIResponse{
		Status:  http.StatusForbidden,
		Message: message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, dto.APIResponse{
		Status:  http.StatusNotFound,
		Message: message,
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, dto.APIResponse{
		Status:  http.StatusConflict,
		Message: message,
	})
}

func Locked(c *gin.Context, message string) {
	c.JSON(http.StatusLocked, dto.APIResponse{
		Status:  http.StatusLocked,
		Message: message,
	})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, dto.APIResponse{
		Status:  http.StatusInternalServerError,
		Message: message,
	})
}

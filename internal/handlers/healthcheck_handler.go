package handlers

import (
	"microsrv/internal/pkg/msresponse"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {

	rsp := gin.H{
		"ok":        true,
		"status":    "up",
		"service":   "microsrv",
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	msresponse.OK(c, http.StatusOK, "sucesso", rsp)
}

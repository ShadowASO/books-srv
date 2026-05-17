package handlers

import (
	"microsrv/internal/pkg/msresponse"
	"microsrv/internal/pkg/mstypes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Versao da aplicação
const AppVersion = "Books 1.2.0"

func VersionHandler(c *gin.Context) {

	rsp := mstypes.JsonMap{
		"version": AppVersion,
	}
	msresponse.OK(c, http.StatusCreated, "Sucesso", rsp)

}

package controller

import "github.com/gin-gonic/gin"

// RestController 方法接口
type RestController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Show(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

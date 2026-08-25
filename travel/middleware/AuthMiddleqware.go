package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"travel/TravelDate"
	"travel/TravelModel"
	"travel/auth"
)

// @title	AuthMiddleware
// /@description	鉴权中间件，鉴定用户权限，鉴定用户是否登录
// @auth	Snactop	2023-11-30	19:54
// @param	无传入参数
// @return	gin.HandlerFunc	返回一个请求处理函数
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//获取 authorization Header
		tokenString := ctx.GetHeader("Authorization")

		//valid token format
		if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "权限不足"})
			ctx.Abort()
			return
		}
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer"))
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "权限不足"})
			ctx.Abort()
			return
		}

		service, err := auth.DefaultService()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": "503", "msg": "认证服务暂不可用"})
			ctx.Abort()
			return
		}
		claim, err := service.ValidateAccess(ctx.Request.Context(), tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "权限不足"})
			ctx.Abort()
			return
		}

		//验证通过获取后的userId
		user := TravelModel.TraUser{}
		userId := claim.UserID
		if userId == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "权限不足"})
			ctx.Abort()
			return
		}

		db := TravelDate.GetDB()
		if err := db.First(&user, userId).Error; err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": "401", "msg": "权限不足"})
			ctx.Abort()
			return
		}

		authInfo := TravelModel.AuthInformation{
			ID:         user.ID,
			OpenID:     user.OpenID,
			SessionKey: user.SessionKey,
		}

		//将用户信息写入上下文
		ctx.Set("authInfo", authInfo)
		ctx.Set("authClaims", claim)
		ctx.Next()
	}
}

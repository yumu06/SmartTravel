package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 路径规划请求体结构
type PlanRouteRequest struct {
	Origin       string   `json:"origin"`       // 起点坐标 "23.429962,116.702396"
	Destinations []string `json:"destinations"` // 一组 "lat,lng" 坐标
	Mode         string   `json:"mode"`         // "driving", "walking", "transit"
}

// 路径规划响应封装（部分）
type RouteResponse struct {
	Status int         `json:"status"`
	Result interface{} `json:"result"`
}

func PlanRoute(c *gin.Context) {
	var req PlanRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数格式错误"})
		return
	}

	// 校验参数
	if req.Origin == "" || len(req.Destinations) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "origin 和 destinations 不能为空"})
		return
	}

	// 默认使用 driving 模式
	if req.Mode == "" {
		req.Mode = "driving"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": req,
		"msg":  "路径规划请由前端使用腾讯位置服务 SDK 完成",
	})
}

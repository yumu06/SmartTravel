package controller

import (
	"net/http"
	"strconv"
	"travel/TravelModel"
	"travel/logic"
	"travel/vo"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func PostCreate(ctx *gin.Context) {
	//参数验证
	var postR vo.PostRequest
	if err := ctx.ShouldBind(&postR); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	//检查用户权限
	authInfo, exit := ctx.Get("authInfo")
	if !exit {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	//逻辑层创建文章，返回文章ID
	postID, err := logic.PostCreate(authInfo.(TravelModel.AuthInformation).ID, postR)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误，文章创建失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "postID": postID, "msg": "文章创建成功"})
	return
}

func PostUpdate(ctx *gin.Context) {
	//参数验证
	var postR vo.PostRequest
	if err := ctx.ShouldBind(&postR); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	//检查用户权限
	authInfo, exit := ctx.Get("authInfo")
	if !exit {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	//在path中寻找文章ID
	postID := ctx.Param("id")

	//用户修改文章信息
	post, err := logic.PostUpdate(authInfo.(TravelModel.AuthInformation).ID, postID, postR)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "文章修改失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "文章信息修改成功", "post": post})
	return
}

func PostShow(ctx *gin.Context) {
	//在path中寻找文章ID
	postID := ctx.Param("id")

	logic.IncrementPostView(postID)

	//逻辑层查找商品信息
	post, err := logic.GetPostInfo(postID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户权限错误", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "文章查询成功", "post": post})
	return
}

func PostDelete(ctx *gin.Context) {
	// 处理商品的删除逻辑
	//在path中寻找商品ID
	postID := ctx.Param("id")

	//检查用户权限
	authInfo, exit := ctx.Get("authInfo")
	if !exit {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	//逻辑层删除商品信息
	if err := logic.PostDelete(authInfo.(TravelModel.AuthInformation).ID, postID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "系统错误", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除文章成功，Delete Post Success"})
}

func PostPageList(ctx *gin.Context) {
	// 处理获取商品列表的逻辑，通常需要分页和筛选
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageList, _ := strconv.Atoi(ctx.DefaultQuery("pageList", "20"))

	posts, total := logic.PageList(pageNum, pageList)

	data := map[string]interface{}{
		"msg":         "商品信息获取成功",
		"commodities": posts,
		"total":       total,
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "date": data, "message": "删除文章成功，Delete Post Success"})
	return
}

func PostSearch(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	posts, total, err := logic.SearchPosts(keyword, pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": posts, "total": total, "msg": "搜索成功"})
}

func PostListByUser(ctx *gin.Context) {
	userIDStr := ctx.Query("userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "userID 参数错误"})
		return
	}
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	posts, total, err := logic.ListPostsByUserID(userID, pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": posts, "total": total, "msg": "获取成功"})
}

func PostRecommend(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	posts, err := logic.RecommendPosts(authInfo.(TravelModel.AuthInformation).ID, limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": posts, "msg": "获取推荐成功"})
}

func PostLike(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	postID := ctx.Param("id")
	if err := logic.AddPostLike(authInfo.(TravelModel.AuthInformation).ID, postID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "点赞成功"})
}

func PostUnlike(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	postID := ctx.Param("id")
	if err := logic.RemovePostLike(authInfo.(TravelModel.AuthInformation).ID, postID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "取消点赞成功"})
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func CommentCreate(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	postID := ctx.Param("id")
	var req CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	commentID, err := logic.CreateComment(authInfo.(TravelModel.AuthInformation).ID, postID, req.Content)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "commentID": commentID, "msg": "评论成功"})
}

func CommentList(ctx *gin.Context) {
	postID := ctx.Param("id")
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	comments, total, err := logic.ListComments(postID, pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": comments, "total": total, "msg": "获取评论成功"})
}

func CommentDelete(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	commentID := ctx.Param("id")
	if err := logic.DeleteComment(authInfo.(TravelModel.AuthInformation).ID, commentID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除评论成功"})
}

type NoticeRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"imageURL"`
	LinkURL  string `json:"linkURL"`
}

func NoticeList(ctx *gin.Context) {
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	notices, total, err := logic.ListNotices(pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": notices, "total": total, "msg": "获取公告成功"})
}

func NoticeShow(ctx *gin.Context) {
	noticeID := ctx.Param("id")
	notice, err := logic.GetNotice(noticeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": notice, "msg": "获取公告成功"})
}

func NoticeCreate(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限"})
		return
	}

	var req NoticeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	noticeID, err := logic.CreateNotice(req.Title, req.Content, req.ImageURL, req.LinkURL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "noticeID": noticeID, "msg": "创建公告成功"})
}

func NoticeUpdate(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限"})
		return
	}

	noticeID := ctx.Param("id")
	var req NoticeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if err := logic.UpdateNotice(noticeID, req.Title, req.Content, req.ImageURL, req.LinkURL); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新公告成功"})
}

func NoticeDelete(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限"})
		return
	}

	noticeID := ctx.Param("id")
	if err := logic.DeleteNotice(noticeID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除公告成功"})
}

func isAdmin(ctx *gin.Context) bool {
	adminToken := viper.GetString("admin.token")
	if adminToken == "" {
		return false
	}
	return ctx.GetHeader("X-Admin-Token") == adminToken
}

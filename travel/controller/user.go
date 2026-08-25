package controller

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"travel/TravelDate"
	"travel/TravelModel"
	"travel/auth"
	"travel/logic"
	traveljwt "travel/pkg/jwt"
	"travel/pkg/snowflake"
	"travel/vo"

	"github.com/gin-gonic/gin"
)

// Login @title  Login
// @description	调用方式：POST； 提交表单额方式：x-www-form-urlencoded；获取微信用户获取用户的openID和SessionKey以计入系统，
// @auth	Snactop	2023-11-27	20:07
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func Login(ctx *gin.Context) {
	db := TravelDate.GetDB()
	wxCode := ctx.PostForm("code")

	//参数验证
	if validInputPattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`); wxCode == "" || !validInputPattern.MatchString(wxCode) {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	SessionKey, OpenID, err := logic.GetIdentify(wxCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err})
		return
	}

	//检查此openID的用户是否注册
	var user TravelModel.TraUser
	db.Where("  open_id  = ?", OpenID).First(&user)
	if user.ID == 0 {
		//用户未注册，为用户注册，创建用户信息
		//TODO 生成用户ID
		userId, err := snowflake.GetID()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "系统繁忙，登录失败"})
			return
		}

		userInfo := TravelModel.TraUser{
			ID:         userId,
			OpenID:     OpenID,
			SessionKey: SessionKey,
		}
		if err := db.Create(&userInfo).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误，用户创建失败"})
			return
		}
		user = userInfo
	} else { //更新用户SessionKey
		if err := db.Model(&user).Update("session_key", SessionKey).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "系统错误，session_key更新失败"})
			return
		}
		user.SessionKey = SessionKey
	}

	service, err := auth.DefaultService()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "认证服务暂不可用"})
		return
	}
	pair, err := service.StartSession(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "认证服务暂不可用"})
		return
	}

	writeLoginSuccess(ctx, pair, SessionKey)
	return
}

func writeLoginSuccess(ctx *gin.Context, pair traveljwt.Pair, sessionKey string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":       200,
		"token":      pair.AccessToken,
		"SessionKey": sessionKey,
		"data":       pair,
		"msg":        "登录成功",
	})
}

func AuthRefresh(ctx *gin.Context) {
	var request vo.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "refresh_token 参数错误"})
		return
	}
	service, err := auth.DefaultService()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "认证服务暂不可用"})
		return
	}
	pair, err := service.Refresh(ctx.Request.Context(), request.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Refresh Token 无效或已失效"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": pair, "msg": "刷新成功"})
}

func AuthLogout(ctx *gin.Context) {
	tokenString := strings.TrimSpace(strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer"))
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Access Token 缺失"})
		return
	}
	service, err := auth.DefaultService()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "认证服务暂不可用"})
		return
	}
	if err := service.Logout(ctx.Request.Context(), tokenString); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Access Token 无效或已失效"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "注销成功"})
}

// @title  GetUserProfile
// @description	获取用户信息,暂时没有作用
// @auth	Snactop	2023-11-27	20:07
// @param	ctx *gin.Context  传入一个上下文（ EncryptedData、 Iv 、SessionKey）
// @return	void	没有返回值
func GetUserProfile(ctx *gin.Context) {
	var identifyCode vo.IdentifyCode
	if err := ctx.ShouldBind(&identifyCode); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	//todo 解密数据
	//var userInfo TravelModel.TraUser
	plainText, err := logic.DecryptUserInfo(identifyCode.SessionKey, identifyCode.EncryptedData, identifyCode.Iv)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "服务器错误"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "plainText": plainText, "msg": "操作成功"})
	return
}

// @title  Update
// @description	获取方式：POST；用户自行更改用户信息
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func Update(ctx *gin.Context) {
	var userP vo.UpdateUserRequest
	if err := ctx.ShouldBind(&userP); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求错误或数据格式错误"})
		return
	}

	//TODO 查找用户的openID
	authInfo, exit := ctx.Get("authInfo")
	if !exit {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	if authInfo.(TravelModel.AuthInformation).OpenID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	//更改用户信息
	if err := logic.UpdateUserInformation(userP, authInfo.(TravelModel.AuthInformation).OpenID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
	return
}

// @title  GetUserInformation
// @description	获取方式：GET；展示用户信息用户信息
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func GetUserInformation(ctx *gin.Context) {
	//TODO 查找用户openID
	authInfo, exit := ctx.Get("authInfo")
	if !exit {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	OpenID := authInfo.(TravelModel.AuthInformation).OpenID
	if OpenID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "登录已过期，请重新登录"})
		return
	}

	//查找用户信息
	user, err := logic.GetUserInformation(OpenID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "information": user, "msg": "获取用户信息成功"})
	return
}

func Authorization(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}
	info := authInfo.(TravelModel.AuthInformation)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"userID": info.ID, "openID": info.OpenID}, "msg": "授权有效"})
}

func UserSearch(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	users, err := logic.SearchUsers(keyword, limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": users, "msg": "搜索成功"})
}

func ChatSend(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}
	var req vo.ChatSendRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}
	messageID, err := logic.SendChatMessage(authInfo.(TravelModel.AuthInformation).ID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "messageID": messageID, "msg": "发送成功"})
}

func ChatList(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	withUserIDStr := ctx.Query("withUserID")
	withUserID, err := strconv.ParseUint(withUserIDStr, 10, 64)
	if err != nil || withUserID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "withUserID 参数错误"})
		return
	}

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "50"))
	msgs, total, err := logic.ListChatMessages(authInfo.(TravelModel.AuthInformation).ID, withUserID, pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": msgs, "total": total, "msg": "获取对话成功"})
}

// AddPostStart @title  AddPostStart
// @description	获取方式：POST；用户收藏文章
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func AddPostStart(ctx *gin.Context) {

	// 获取当前登录的用户ID（假设是通过JWT或Session获取的）
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取文章ID
	postID := ctx.Param("id") // 从URL中获取 /users/favorites/:id
	if len(postID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	// 逻辑层收藏商品
	code, err := logic.AddFavoritePost(authInfo.(TravelModel.AuthInformation).ID, postID)
	if err != nil {
		ctx.JSON(code, gin.H{"message": "收藏失败", "error": err.Error()})
		return
	}

	ctx.JSON(code, gin.H{"message": "收藏成功"})
	return

}

// RemovePostStart @title  RemovePostStart
// @description	获取方式：GET；用户删除收藏文章
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func RemovePostStart(ctx *gin.Context) {
	// 获取当前登录的用户ID
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取商品ID
	postID := ctx.Param("id")
	if len(postID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	// 删除用户的收藏记录
	code, err := logic.RemovePostStart(authInfo.(TravelModel.AuthInformation).ID, postID)
	if err != nil {
		ctx.JSON(code, gin.H{"message": "取消收藏失败", "error": err.Error()})
		return
	}

	ctx.JSON(code, gin.H{"message": "取消收藏成功"})
	return
}

func GetPostStart(ctx *gin.Context) {
	// 获取当前登录的用户ID
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	//获取用户收藏列表
	posts, err := logic.GetUserPostStart(authInfo.(TravelModel.AuthInformation).ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "获取收藏失败", "err": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": posts, "msg": "获取用户收藏列表成功", "err": nil})
	return
}

func GetUserCreatedPosts(ctx *gin.Context) {
	// 获取当前登录的用户ID
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	posts, err := logic.GetUserCreatedPosts(authInfo.(TravelModel.AuthInformation).ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "获取收藏失败", "err": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": posts, "msg": "获取用户创建商品列表成功", "err": nil})
	return
}

// @title  Favorite
// @description	获取方式：POST；用户收藏足迹
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func FavoriteF(ctx *gin.Context) {
	AddFootStart(ctx)
}

// @title  ShowUserFoot
// @description	获取方式：GET；展示用户历史足迹
// @auth	Snactop	2024-10-12	0:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func ShowUserFoot(ctx *gin.Context) {
	FootList(ctx)
}

type CreateFootByPlanRouteRequest struct {
	Title            string          `json:"title"`
	Origin           string          `json:"origin"`
	OriginName       string          `json:"origin_name"`
	Destinations     []string        `json:"destinations"`
	DestinationNames []string        `json:"destination_names"`
	Mode             string          `json:"mode"`
	RouteResult      json.RawMessage `json:"routeResult"`
}

func FootCreate(ctx *gin.Context) {
	var req CreateFootByPlanRouteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数格式错误"})
		return
	}
	if req.Origin == "" || len(req.Destinations) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "origin 和 destinations 不能为空"})
		return
	}
	if req.Mode == "" {
		req.Mode = "driving"
	}
	if len(req.RouteResult) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "routeResult 不能为空"})
		return
	}

	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	destsJSON, _ := json.Marshal(req.Destinations)
	destNamesJSON, _ := json.Marshal(req.DestinationNames)
	footID, err := logic.CreateFoot(authInfo.(TravelModel.AuthInformation).ID, vo.CreateFootRequest{
		Title:            req.Title,
		Origin:           req.Origin,
		OriginName:       req.OriginName,
		Destinations:     string(destsJSON),
		DestinationNames: string(destNamesJSON),
		Mode:             req.Mode,
		RouteResult:      string(req.RouteResult),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "footID": footID, "routeResult": req.RouteResult, "msg": "足迹创建成功"})
}

func FootList(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	foots, total, err := logic.GetUserFootList(authInfo.(TravelModel.AuthInformation).ID, pageNum, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": foots, "total": total, "msg": "获取足迹列表成功"})
}

func FootShow(ctx *gin.Context) {
	footID := ctx.Param("id")
	foot, err := logic.GetFootDetail(footID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": foot, "msg": "获取足迹成功"})
}

func AddFootStart(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	footID := ctx.Param("id")
	if footID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的足迹ID"})
		return
	}

	code, err := logic.AddFavoriteFoot(authInfo.(TravelModel.AuthInformation).ID, footID)
	if err != nil {
		ctx.JSON(code, gin.H{"code": code, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "收藏成功"})
}

func RemoveFootStart(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	footID := ctx.Param("id")
	if footID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的足迹ID"})
		return
	}

	code, err := logic.RemoveFavoriteFoot(authInfo.(TravelModel.AuthInformation).ID, footID)
	if err != nil {
		ctx.JSON(code, gin.H{"code": code, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "取消收藏成功"})
}

func GetFootStart(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}

	foots, err := logic.GetUserFootStart(authInfo.(TravelModel.AuthInformation).ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": foots, "msg": "获取足迹收藏列表成功"})
}

func FootDelete(ctx *gin.Context) {
	authInfo, exists := ctx.Get("authInfo")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录"})
		return
	}
	footID := ctx.Param("id")
	if footID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的足迹ID"})
		return
	}
	if err := logic.DeleteFoot(authInfo.(TravelModel.AuthInformation).ID, footID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除足迹成功"})
}

// @title  ShowP
// @description	获取方式：GET；展示用户收藏的文章信息
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func ShowP(ctx *gin.Context) {
	//TODO 获取用户ID

	//TODO 展示用户收藏的文章列表（要专门写一个文章列表的格式）
}

// @title  ShowF
// @description	获取方式：GET；展示用户收藏的足迹信息
// @auth	Snactop	2024-9-20	15:13
// @param	ctx *gin.Context  传入一个上下文
// @return	void	没有返回值
func ShowF(ctx *gin.Context) {
	//TODO 获取用户ID

	//TODO 展示用户收藏的足迹列表（要专门写一个足迹列表的格式）
}

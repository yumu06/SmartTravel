package router

import (
	"travel/controller"
	"travel/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine) *gin.Engine {
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.POST("/travel/login", controller.Login)
	r.POST("/travel/auth/refresh", controller.AuthRefresh)
	r.POST("/travel/GetUserProfile", controller.GetUserProfile)
	r.GET("/travel/notice/list", controller.NoticeList)
	r.GET("/travel/notice/show/:id", controller.NoticeShow)

	vi := r.Group("/travel")
	vi.Use(middleware.AuthMiddleware()) // 应用JWT认证中间件
	{
		vi.POST("/auth/logout", controller.AuthLogout)
		vi.GET("/authorization", controller.Authorization)
		vi.GET("/recommend", controller.Recommend)
		vi.POST("/planRoute", controller.PlanRoute)

		AdminNoticeControllers := vi.Group("/admin/notice")
		AdminNoticeControllers.POST("/create", controller.NoticeCreate)
		AdminNoticeControllers.PATCH("/update/:id", controller.NoticeUpdate)
		AdminNoticeControllers.DELETE("/delete/:id", controller.NoticeDelete)

		FootControllers := vi.Group("/foot")
		FootControllers.POST("/create", controller.FootCreate)
		FootControllers.GET("/list", controller.FootList)
		FootControllers.GET("/show/:id", controller.FootShow)
		FootControllers.DELETE("/delete/:id", controller.FootDelete)

		FootStartControllers := vi.Group("/foot/start")
		FootStartControllers.POST("/add/:id", controller.AddFootStart)
		FootStartControllers.DELETE("/remove/:id", controller.RemoveFootStart)
		FootStartControllers.GET("/list", controller.GetFootStart)

		//用户
		UserControllers := vi.Group("/user")
		UserControllers.PATCH("/update", controller.Update)
		UserControllers.GET("/info", controller.GetUserInformation)
		UserControllers.GET("/postCreate", controller.GetUserCreatedPosts)
		UserControllers.GET("/search", controller.UserSearch)
		UserControllers.GET("/chat", controller.ChatList)
		UserControllers.POST("/chat", controller.ChatSend)

		//用户收藏文章
		UserStartControllers := vi.Group("/user/start")
		UserStartControllers.POST("/add/:id", controller.AddPostStart)
		UserStartControllers.DELETE("/remove/:id", controller.RemovePostStart)
		UserStartControllers.GET("/list", controller.GetPostStart) // TODO 获取用户收藏列表，这里和PageList有点像，但是这里好像有点问题，之后解决一下

		//文章路由
		PostRouters := vi.Group("/post")
		PostRouters.POST("/create", controller.PostCreate)
		PostRouters.PATCH("/update/:id", controller.PostUpdate)
		PostRouters.GET("/show/:id", controller.PostShow)
		PostRouters.DELETE("/delete/:id", controller.PostDelete)
		PostRouters.GET("/page/list", controller.PostPageList)
		PostRouters.GET("/search", controller.PostSearch)
		PostRouters.GET("/user", controller.PostListByUser)
		PostRouters.GET("/recommand", controller.PostRecommend)
		PostRouters.POST("/like/:id", controller.PostLike)
		PostRouters.DELETE("/like/:id", controller.PostUnlike)
		PostRouters.POST("/:id/comment", controller.CommentCreate)
		PostRouters.GET("/:id/comment/list", controller.CommentList)
		PostRouters.DELETE("/comment/:id", controller.CommentDelete)

	}

	legacy := r.Group("/")
	legacy.Use(middleware.AuthMiddleware())
	{
		legacy.POST("/user/start/add/:id", controller.AddPostStart)
		legacy.DELETE("/user/start/remove/:id", controller.RemovePostStart)
		legacy.GET("/user/start/list", controller.GetPostStart)
	}

	return r
}

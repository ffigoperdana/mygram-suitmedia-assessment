package router

import (
	"finalproject/config"
	"finalproject/controllers"
	"finalproject/middlewares"
	"finalproject/models"
	"time"

	"github.com/gin-gonic/gin"

	_ "finalproject/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Mygram API
// @version 1.0
// @description This is a final project API from Hactiv8 to add photos, comments, and store the social media of users
// @termsOfService http://swagger.io/terms
// @contact.name API Support
// @contact.email perdanaputrafigo@gmail.com
// @license.name Apache 2.0
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @license.url http://www.apache.org/licenses/license-2.0.html
// @BasePath /
func StartApp() *gin.Engine {
	cfg := config.Load()
	r := gin.Default()
	r.MaxMultipartMemory = cfg.S3UploadMaxBytes
	r.Use(middlewares.SecurityHeaders())
	r.Use(middlewares.CORS())

	registerHealthRoutes(r)
	registerMediaRoutes(r)
	registerLegacyRoutes(r, cfg)
	registerV1Routes(r.Group("/api/v1"), cfg)
	registerDocsRoutes(r, cfg)

	return r
}

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/health", controllers.HealthCheck)
	r.GET("/health/ready", controllers.ReadinessCheck)
	r.GET("/health/live", controllers.LivenessCheck)
}

func registerMediaRoutes(r *gin.Engine) {
	r.GET("/media/*objectKey", controllers.ServeMediaObject)
	r.HEAD("/media/*objectKey", controllers.ServeMediaObject)
}

func registerLegacyRoutes(r *gin.Engine, cfg config.Config) {
	standardLimit := middlewares.RedisRateLimit(
		"legacy-api",
		cfg.RateLimitRequests,
		time.Duration(cfg.RateLimitWindowSeconds)*time.Second,
	)

	userRouter := r.Group("/users")
	userRouter.Use(middlewares.RedisRateLimit(
		"legacy-auth",
		cfg.AuthRateLimitRequests,
		time.Duration(cfg.RateLimitWindowSeconds)*time.Second,
	))
	{
		userRouter.POST("/register", controllers.UserRegister)
		userRouter.POST("/login", controllers.UserLogin)
	}

	photoRouter := r.Group("/photos")
	{
		photoRouter.Use(standardLimit, middlewares.Authentication())
		photoRouter.POST("/create", controllers.CreatePhoto)
		photoRouter.GET("/getall", controllers.GetAllPhotos)
		photoRouter.GET("/get/:photoId", controllers.GetPhoto)
		photoRouter.PUT("/update/:photoId", middlewares.PhotoAuthorization(), controllers.UpdatePhoto)
		photoRouter.DELETE("/delete/:photoId", middlewares.PhotoAuthorization(), controllers.DeletePhoto)
	}

	commentRouter := r.Group("/comments")
	{
		commentRouter.Use(standardLimit, middlewares.Authentication())
		commentRouter.POST("/create/:photoId", controllers.CreateComment)
		commentRouter.GET("/getall", controllers.GetAllComments)
		commentRouter.GET("/getall/:photoId", controllers.GetAllCommentsForPhoto)
		commentRouter.GET("/get/:commentId", controllers.GetComment)
		commentRouter.PUT("/update/:commentId", middlewares.CommentAuthorization(), controllers.UpdateComment)
		commentRouter.DELETE("/delete/:commentId", middlewares.CommentAuthorization(), controllers.DeleteComment)
	}

	socmedRouter := r.Group("/socialmedia")
	{
		socmedRouter.Use(standardLimit, middlewares.Authentication())
		socmedRouter.POST("/create", controllers.CreateSocialMedia)
		socmedRouter.GET("/getall", controllers.GetAllSocialMedias)
		socmedRouter.GET("/get/:socialMediaId", controllers.GetSocialMedia)
		socmedRouter.PUT("/update/:socialMediaId", middlewares.SocialMediaAuthorization(), controllers.UpdateSocialMedia)
		socmedRouter.DELETE("/delete/:socialMediaId", middlewares.SocialMediaAuthorization(), controllers.DeleteSocialMedia)
	}
}

func registerV1Routes(api *gin.RouterGroup, cfg config.Config) {
	api.Use(middlewares.RedisRateLimit(
		"api-v1",
		cfg.RateLimitRequests,
		time.Duration(cfg.RateLimitWindowSeconds)*time.Second,
	))

	authRouter := api.Group("/auth")
	authRouter.Use(middlewares.RedisRateLimit(
		"auth-v1",
		cfg.AuthRateLimitRequests,
		time.Duration(cfg.RateLimitWindowSeconds)*time.Second,
	))
	{
		authRouter.POST("/register", controllers.UserRegister)
		authRouter.POST("/login", controllers.UserLogin)
	}

	meRouter := api.Group("/me")
	{
		meRouter.Use(middlewares.Authentication())
		meRouter.GET("", controllers.GetMe)
		meRouter.PATCH("", controllers.UpdateMe)
	}

	photoRouter := api.Group("/photos")
	{
		photoRouter.Use(middlewares.Authentication())
		photoRouter.POST("", controllers.CreatePhoto)
		photoRouter.GET("", controllers.GetAllPhotos)
		photoRouter.GET("/:photoId", controllers.GetPhoto)
		photoRouter.PUT("/:photoId", middlewares.PhotoAuthorization(), controllers.UpdatePhoto)
		photoRouter.DELETE("/:photoId", middlewares.PhotoAuthorization(), controllers.DeletePhoto)
	}

	uploadRouter := api.Group("/uploads")
	{
		uploadRouter.Use(middlewares.Authentication())
		uploadRouter.POST("/photos", controllers.UploadPhotoImage)
	}

	pushRouter := api.Group("/push")
	{
		pushRouter.Use(middlewares.Authentication())
		pushRouter.GET("/vapid-public-key", controllers.GetPushVAPIDPublicKey)
		pushRouter.POST("/subscriptions", controllers.SavePushSubscription)
		pushRouter.DELETE("/subscriptions", controllers.DeletePushSubscription)
	}

	commentRouter := api.Group("/comments")
	{
		commentRouter.Use(middlewares.Authentication())
		commentRouter.GET("", controllers.GetAllComments)
		commentRouter.GET("/:commentId", controllers.GetComment)
		commentRouter.PUT("/:commentId", middlewares.CommentAuthorization(), controllers.UpdateComment)
		commentRouter.DELETE("/:commentId", middlewares.CommentAuthorization(), controllers.DeleteComment)
	}

	photoCommentsRouter := api.Group("/photos/:photoId/comments")
	{
		photoCommentsRouter.Use(middlewares.Authentication())
		photoCommentsRouter.GET("", controllers.GetAllCommentsForPhoto)
		photoCommentsRouter.POST("", controllers.CreateComment)
	}

	socialRouter := api.Group("/social-media")
	{
		socialRouter.Use(middlewares.Authentication())
		socialRouter.POST("", controllers.CreateSocialMedia)
		socialRouter.GET("", controllers.GetAllSocialMedias)
		socialRouter.GET("/:socialMediaId", controllers.GetSocialMedia)
		socialRouter.PUT("/:socialMediaId", middlewares.SocialMediaAuthorization(), controllers.UpdateSocialMedia)
		socialRouter.DELETE("/:socialMediaId", middlewares.SocialMediaAuthorization(), controllers.DeleteSocialMedia)
	}

	adminRouter := api.Group("/admin")
	{
		adminRouter.Use(middlewares.Authentication(), middlewares.RequireRole(models.RoleAdmin))
		adminRouter.GET("/stats", controllers.AdminStats)
		adminRouter.GET("/users", controllers.AdminListUsers)
		adminRouter.GET("/users/:userId", controllers.AdminGetUser)
		adminRouter.PATCH("/users/:userId", controllers.AdminUpdateUser)
		adminRouter.DELETE("/users/:userId", controllers.AdminDeleteUser)
		adminRouter.POST("/users/:userId/ban", controllers.AdminBanUser)
		adminRouter.POST("/users/:userId/unban", controllers.AdminUnbanUser)
	}
}

func registerDocsRoutes(r *gin.Engine, cfg config.Config) {
	if cfg.PublicOpenAPI || cfg.SwaggerUIMode == "public" {
		r.GET("/openapi/public.json", controllers.PublicOpenAPISpec)
	}

	switch cfg.SwaggerUIMode {
	case "internal":
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	case "public":
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerfiles.Handler,
			ginSwagger.URL("/openapi/public.json"),
		))
	}
}

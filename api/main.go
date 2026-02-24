package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"tiehfood/thrifty/docs"
)

func main() {
	dbCon, err := initAndOpenDb()
	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	docs.SwaggerInfo.Title = "Thrifty API"
	docs.SwaggerInfo.Description = "This is the documentation for the Thrifty API"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Version = "1.0"

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PATCH", "OPTIONS", "DELETE"},
		AllowHeaders: []string{"Content-Type"},
	}))

	router.Use(func(context *gin.Context) {
		context.Set("dbCon", dbCon)
		context.Next()
	})

	v1 := router.Group("/api")
	{
		flows := v1.Group("/flows")
		{
			flows.GET("", getFlows)
			flows.POST("", addFlow)
			flows.PATCH(":id", updateFlow)
			flows.DELETE(":id", deleteFlow)
		}
		users := v1.Group("/users")
		{
			users.GET("", getUsers)
			users.POST("", createUser)
			users.PATCH(":id", updateUser)
			users.DELETE(":id", deleteUser)
		}
		settings := v1.Group("/settings")
		{
			settings.GET("", getSettings)
			settings.PATCH("", updateSettings)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port, err := getAndValidatePort()
	if err != nil {
		fmt.Println(errorPrefix, err)
	}
	fmt.Printf("Running on port: %d\n", port)
	err = router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(errorPrefix, err)
	}
}

func getAndValidatePort() (int, error) {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080, nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8080, fmt.Errorf("invalid PORT value: %v", err)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("PORT value out of range (1-65535)")
	}

	return port, nil
}

package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Coreychen4444/shortvideo_ms-api_gateway/auth"
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/handler"
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/routers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	auth.Create_crt_key()
	time.Sleep(1 * time.Second)
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	UC := handler.UserClient()
	VC := handler.VideoClient()
	rHttps := routers.InitRouter(UC, VC)
	go func() {
		rHttps.RunTLS(":443", os.Getenv("CRT"), os.Getenv("KEY"))
	}()
	//http 请求
	rHttp := gin.Default()
	rHttp.Any("/*any", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.URL.RequestURI())
	})
	rHttp.Run(":80")
}

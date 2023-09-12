package main

import (
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/handler"
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/routers"
)

func main() {
	UC := handler.UserClient()
	VC := handler.VideoClient()
	r := routers.InitRouter(UC, VC)
	r.Run(":8080")
}

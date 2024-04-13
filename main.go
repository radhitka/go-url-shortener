package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ShortRequest struct {
	Url string `form:"url"`
}

type DataUrls struct {
	urls map[string]string
}

func main() {

	dataUrls := &DataUrls{
		urls: make(map[string]string),
	}

	route := gin.Default()

	route.GET("/short/:code", dataUrls.getShort)
	route.POST("/short", dataUrls.postShort)

	route.NoRoute(func(ctx *gin.Context) {
		res := &ResponseData{
			Code:    404,
			Message: "Route Not Found!",
		}

		ctx.JSON(res.Code, res)
	})

	route.Run("localhost:8080")
}

func (du *DataUrls) getShort(ctx *gin.Context) {
	code := ctx.Param("code")

	if code == "" {
		res := &ResponseData{
			Code:    http.StatusBadRequest,
			Message: "Code is missing!",
		}

		ctx.JSON(res.Code, res)
		return
	}

	urlOri, found := du.urls[code]

	if !found {
		res := &ResponseData{
			Code:    http.StatusNotFound,
			Message: "Code not found!",
		}

		ctx.JSON(res.Code, res)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, urlOri)
}

func (du *DataUrls) postShort(ctx *gin.Context) {
	var shortRequest ShortRequest

	ctx.Bind(&shortRequest)

	shortKey := getRandKey()

	du.urls[shortKey] = shortRequest.Url

	res := &ResponseData{
		Code:    200,
		Message: "Success",
		Data: map[string]any{
			"urlORi":   shortRequest.Url,
			"urlShort": fmt.Sprintf("http://localhost:8080/short/%s", shortKey),
		},
	}

	ctx.JSON(res.Code, res)
}

func getRandKey() string {

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]

	}

	fmt.Println(charset[1])

	return string(shortKey)

}

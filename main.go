package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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
	urls   map[string]string
	Redist *redis.Client
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error load Env file")
	}

	redis := initRedis()

	defer redis.Close()

	dataUrls := &DataUrls{
		urls:   make(map[string]string),
		Redist: redis,
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

	Newredis := du.Redist

	data, _ := Newredis.Get(ctx, code).Result()

	defer Newredis.Close()

	if data == "" {
		res := &ResponseData{
			Code:    http.StatusNotFound,
			Message: "Code not found!",
		}

		ctx.JSON(res.Code, res)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, data)
}

func (du *DataUrls) postShort(ctx *gin.Context) {

	Newredis := du.Redist

	//limiter
	val, err := Newredis.Get(ctx, ctx.ClientIP()).Result()
	limit, _ := Newredis.TTL(ctx, ctx.ClientIP()).Result()

	if err == redis.Nil {
		Newredis.Set(ctx, ctx.ClientIP(), 10, time.Minute*2)
	} else if err == nil {

		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			ctx.JSON(http.StatusServiceUnavailable, map[string]any{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
			return
		}

	}

	var shortRequest ShortRequest

	ctx.Bind(&shortRequest)

	shortKey := getRandKey()

	du.urls[shortKey] = shortRequest.Url

	dataCode, _ := Newredis.Get(ctx, shortKey).Result()

	if dataCode != "" {
		ctx.JSON(http.StatusForbidden, map[string]any{
			"error": "URL Custom short is already in use",
		})
		return
	}

	err = Newredis.Set(ctx, shortKey, shortRequest.Url, 24*3600*time.Second).Err()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "Unable to connect to server",
			"message": err.Error(),
		})
		return
	}

	remainingQuota, _ := Newredis.Decr(ctx, ctx.ClientIP()).Result()

	res := &ResponseData{
		Code:    200,
		Message: "Success",
		Data: map[string]any{
			"urlORi":           shortRequest.Url,
			"urlShort":         fmt.Sprintf("http://localhost:8080/short/%s", shortKey),
			"rate_limit":       int(remainingQuota),
			"rate_limit_reset": int(limit / time.Nanosecond / time.Minute),
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

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIST_ADD"),
		Password: os.Getenv("REDIST_PASS"),
	})

	return client
}

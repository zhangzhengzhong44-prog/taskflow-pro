package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"taskflow-pro/backend/internal/response"
	"taskflow-pro/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return param.TimeStamp.Format(time.RFC3339) + " " +
			param.Method + " " + param.Path + " " +
			strconv.Itoa(param.StatusCode) + " " +
			param.Latency.String() + "\n"
	})
}

func RateLimit(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := "rate:" + c.ClientIP()
		count, err := redis.Incr(ctx, key).Result()
		if err == nil && count == 1 {
			_ = redis.Expire(ctx, key, time.Minute).Err()
		}
		if err == nil && count > 120 {
			response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, http.StatusUnauthorized, "请先登录")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	value, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	userID, _ := value.(uint)
	return userID
}

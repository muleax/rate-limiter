package main

import (
	"context"
	"log"
	"time"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
)

var ctx = context.Background()

type RequestHandler struct {
	RedisClinet	*redis.Client
	AppClient	fasthttp.Client	
}

func (h RequestHandler) Handle(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-KEY");
	if apiKey == "" {
		return c.Status(http.StatusBadRequest).SendString("missing api key (X-API-KEY)\n")
	}

	pipe := h.RedisClinet.TxPipeline()

	pipe.SetNX(ctx, apiKey, 0, 5 * time.Second)
	incr := pipe.Incr(ctx, apiKey)

	if _, err := pipe.Exec(ctx);  err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error() + "\n")
	}
	
	if (incr.Val() > 3) {
		return c.Status(http.StatusTooManyRequests).SendString("rate limit exceeded\n")
	}

	req := c.Request()
	
	originalURL := utils.CopyString(c.OriginalURL())
	defer req.SetRequestURI(originalURL)
	
	req.URI().SetHost("app-server:8000")

	// redisClinet.Header.Del(fiber.HeaderConnection)
	return h.AppClient.Do(req, c.Response())
}

func main() {
	time.Sleep(3 * time.Second)

	handler := new(RequestHandler)

	handler.RedisClinet = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	handler.AppClient = fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
	}

	app := fiber.New()

    app.Use(handler.Handle)

    log.Print(app.Listen(":8080"))
}

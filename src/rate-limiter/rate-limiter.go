package main

import (
    "context"
    "log"
    "flag"
    "time"
    "net/http"

    "github.com/go-redis/redis/v9"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/utils"
    "github.com/valyala/fasthttp"
)

var ctx = context.Background()

type RequestHandler struct {
    RedisClinet *redis.Client
    AppClient   fasthttp.Client
    AppEndpoint string
    Window      uint
    Limit       uint
}

func (h RequestHandler) Handle(c *fiber.Ctx) error {
    if h.Limit > 0 {
        apiKey := c.Get("X-API-KEY");
        if apiKey == "" {
            return c.Status(http.StatusBadRequest).SendString("missing api key (X-API-KEY)")
        }

        pipe := h.RedisClinet.TxPipeline()

        pipe.SetNX(ctx, apiKey, 0, time.Duration(h.Window) * time.Second)
        incr := pipe.Incr(ctx, apiKey)

        if _, err := pipe.Exec(ctx);  err != nil {
            return c.Status(http.StatusInternalServerError).SendString(err.Error())
        }
        
        if (incr.Val() > int64(h.Limit)) {
            return c.Status(http.StatusTooManyRequests).SendString("rate limit exceeded")
        }
    }

    req := c.Request()
    
    originalURL := utils.CopyString(c.OriginalURL())
    defer req.SetRequestURI(originalURL)
    
    req.URI().SetHost(h.AppEndpoint)

    // redisClinet.Header.Del(fiber.HeaderConnection)
    return h.AppClient.Do(req, c.Response())
}

func main() {
    window          := flag.Uint("window", 5, "window size, sec")
    limit           := flag.Uint("limit", 5, "request limit per token per window")
    appEndpoint     := flag.String("app-endpoint", "app-server:8080", "app service endpoint")
    redisEndpoint   := flag.String("redis-endpoint", "redis:6379", "Redis db endpoint")
    port            := flag.String("port", "8080", "service port")

    flag.Parse()

    handler := RequestHandler{
        Window:         *window,
        Limit:          *limit,
        AppEndpoint:    *appEndpoint,
    }

    // waiting for Redis
    // TODO: figure out better workaround
    time.Sleep(3 * time.Second)

    handler.RedisClinet = redis.NewClient(&redis.Options{
        Addr:     *redisEndpoint,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    handler.AppClient = fasthttp.Client{
        NoDefaultUserAgentHeader: true,
        DisablePathNormalizing:   true,
    }

    app := fiber.New()

    app.Use(handler.Handle)

    log.Print(app.Listen(":" + *port))
}

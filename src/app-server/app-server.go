package main

import (
    "flag"
    "log"
    "net/http"
    "strconv"

    "github.com/gofiber/fiber/v2"
)

type BinaryOpParams struct {
    Lhs float64
    Rhs float64
}

type BinaryOp func(float64, float64) float64

func applyBinaryOp(c *fiber.Ctx, op BinaryOp) (string, error) {
    params := new(BinaryOpParams)
    
    if err := c.QueryParser(params); err != nil {
        return "", err
    }
    
    value := op(params.Lhs, params.Rhs)
    converted := strconv.FormatFloat(value, 'g', -1, 64)

    return converted, nil
} 

func main() {
    port := flag.String("port", "8080", "service port")
    
    flag.Parse()

    app := fiber.New()

    // GET /api/add
    app.Get("/api/add", func(c *fiber.Ctx) error {
        ans, err := applyBinaryOp(c, func(lhs float64, rhs float64) float64 { return lhs + rhs })
        if (err != nil) {
            return c.Status(http.StatusBadRequest).SendString(err.Error())
        }
        return c.SendString(ans)
    })

    // GET /api/sub
    app.Get("/api/sub", func(c *fiber.Ctx) error {
        ans, err := applyBinaryOp(c, func(lhs float64, rhs float64) float64 { return lhs - rhs })
        if (err != nil) {
            return c.Status(http.StatusBadRequest).SendString(err.Error())
        }
        return c.SendString(ans)
    })

    // GET /api/mul
    app.Get("/api/sub", func(c *fiber.Ctx) error {
        ans, err := applyBinaryOp(c, func(lhs float64, rhs float64) float64 { return lhs * rhs })
        if (err != nil) {
            return c.Status(http.StatusBadRequest).SendString(err.Error())
        }
        return c.SendString(ans)
    })

    // GET /api/div
    app.Get("/api/div", func(c *fiber.Ctx) error {
        ans, err := applyBinaryOp(c, func(lhs float64, rhs float64) float64 { return lhs / rhs })
        if (err != nil) {
            return c.Status(http.StatusBadRequest).SendString(err.Error())
        }
        return c.SendString(ans)
    })

    log.Fatal(app.Listen(":" + *port))
}

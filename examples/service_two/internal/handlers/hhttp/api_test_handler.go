package hhttp

import "github.com/gofiber/fiber/v2"

type ApiTestHandler struct {
}

func NewApiTestHandler() *ApiTestHandler {
	return &ApiTestHandler{}
}

func (ath *ApiTestHandler) HandleApiTestV1(c *fiber.Ctx) error {
	return c.SendString("good")
}

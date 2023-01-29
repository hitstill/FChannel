package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetTheme(ctx *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "theme"
	cookie.Value = ctx.FormValue("theme")
	ctx.Cookie(cookie)
	return ctx.RedirectBack("/")
}

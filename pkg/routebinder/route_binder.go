package routebinder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lesismal/arpc"
)

type FiberRouteBinder interface {
	BindFiber(fa *fiber.App)
}

type ArpcRouterBinder interface {
	BindArpc(srv *arpc.Server)
}

type AppRouterBinder interface {
	FiberRouteBinder

	ArpcRouterBinder
}

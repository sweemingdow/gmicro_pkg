package gserver

import (
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/pkg/utils"
	"strconv"
)

const (
	pathRewriteType = "path_rewrite"
)

type MatchRule struct {
	Type string
	Path string
	Args map[string]any
}

// Match Rule create factory
func createRuleHandler(srvName string, mr MatchRule) fiber.Handler {
	if mr.Type == pathRewriteType {
		return pathRewriteHandler(srvName, mr.Args)
	}

	return nil
}

var emptyInfo = GwMetaInfo{}

type GwMetaInfo struct {
	Id         string // which upstream server
	RoutedPath string // finally reached path
	ReqId      string // req-id
}

type metaInfoCtxKey struct {
}

func GetMetaInfoFromCtx(c *fiber.Ctx) GwMetaInfo {
	if info, ok := c.Locals(metaInfoCtxKey{}).(GwMetaInfo); ok {
		return info
	}

	return emptyInfo
}

func pathRewriteHandler(srvName string, args map[string]any) fiber.Handler {
	depth := getDepthFromArgs(args)
	if depth == 0 {
		return nil
	}
	// /endpoints/api/one_service/**
	// /endpoints/api/one_service/test/v1
	// /test/v1

	return func(c *fiber.Ctx) error {
		req := c.Request()
		uri := req.URI()

		pathBytes := uri.Path()

		var (
			curIdx   = 0
			curDepth = 0
		)

		for i := 1; i < len(pathBytes) && curDepth < depth; i++ {
			if pathBytes[i] == '/' {
				curDepth++
				curIdx = i
			}
		}

		newPathBytes := uri.Path()[curIdx:]
		if len(newPathBytes) == 0 {
			newPathBytes = []byte{'/'}
		}

		uri.SetPathBytes(newPathBytes)

		// store which upstream server had be proxy. may can use it for statistics
		c.Locals(
			metaInfoCtxKey{},
			GwMetaInfo{
				Id:         srvName,
				RoutedPath: string(newPathBytes),
				ReqId:      utils.RandStr(32),
			},
		)

		return nil
	}
}

func getDepthFromArgs(args map[string]any) int {
	val := args["depth"]

	if f64, ok := val.(float64); ok {
		return int(f64)
	}

	if is, ok := val.(string); ok {
		v, err := strconv.Atoi(is)
		if err != nil {
			return 0
		}

		return v
	}

	if i, ok := val.(int); ok {
		return i
	}

	return 0
}

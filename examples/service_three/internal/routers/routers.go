package routers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lesismal/arpc"
	"gmicro_pkg/pkg/routebinder"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type threeServiceRouteBinder struct {
	hCli *http.Client
}

func NewThreeServiceRouteBinder() routebinder.AppRouterBinder {
	return &threeServiceRouteBinder{
		hCli: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:       512,
				MaxIdleConnsPerHost:   512,
				MaxIdleConns:          512,
				IdleConnTimeout:       90 * time.Second,
				ResponseHeaderTimeout: 60 * time.Second,
				DisableKeepAlives:     false,
			},
		},
	}
}

func (rb *threeServiceRouteBinder) BindFiber(fa *fiber.App) {
	fa.Get("/test/v3", func(c *fiber.Ctx) error {
		guc := c.QueryInt("guc", 10)
		cnt := c.QueryInt("cnt", 100)

		var errGrp errgroup.Group

		errGrp.SetLimit(guc)

		start := time.Now()

		var counter, errCounter atomic.Uint32
		var lastErr atomic.Value

		for j := 0; j < cnt; j++ {
			errGrp.Go(func() error {
				err := rb.sendReq()
				if err != nil {
					errCounter.Add(1)
					lastErr.Store(err)
				}

				counter.Add(1)
				return nil
			})
		}

		_ = errGrp.Wait()
		info := fmt.Sprintf("gateway test completed, guc:%d, cnt:%d, doneCnt:%d, errCnt:%d, lastErr:%v, took:%v",
			guc, cnt, counter.Load(), errCounter.Load(), lastErr.Load(), time.Since(start))

		log.Println(info)

		return c.SendString(info)
	})

}

func (rb *threeServiceRouteBinder) sendReq() error {
	//val := utils.RandInt(1, 10)
	var host string
	//if val <= 7 {
	//	host = "http://localhost:6010/endpoints/api/one_service/test/v1"
	//} else {
	//	host = "http://localhost:6010/endpoints/api/two_service/test/v2"
	//}

	host = "http://localhost:6010/endpoints/api/one_service/test/v1"

	resp, err := rb.hCli.Get(host)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("http status:%d not ok", resp.StatusCode))
	}

	return nil
}

func (rb *threeServiceRouteBinder) BindArpc(srv *arpc.Server) {

}

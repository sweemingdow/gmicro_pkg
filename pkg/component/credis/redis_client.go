package credis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"sync/atomic"
	"time"
)

type RedisCfg struct {
	Addresses      []string // [192.168.1.1:6379]
	Database       int
	Username       string
	Password       string
	MinimumIdle    int           // 最小限制连接数
	MaximumIdle    int           // 最大闲置连接数
	Maximum        int           // 池中最大连接数
	MaxIdleTime    time.Duration // 最大闲置时间
	ReadTimeout    time.Duration // 读取超时
	WriteTimeout   time.Duration // 写超时
	MaxWaitTimeout time.Duration // 连接满时, 最大的等待时间
	PingTimeout    time.Duration // ping超时
}

type RedisClient struct {
	cli    redis.UniversalClient
	cfg    RedisCfg
	closed atomic.Bool
}

func (rc *RedisClient) OnCreated(ec chan<- error) {
	ctx, cancel := context.WithTimeout(context.Background(), rc.cfg.PingTimeout)
	defer cancel()

	if err := rc.cli.Ping(ctx).Err(); err != nil {
		ec <- err
		return
	}

	lg := mylog.AppLogger()
	lg.Info().Msg("int redis client successfully")
}

func (rc *RedisClient) OnDispose(ctx context.Context) error {
	if !rc.closed.CompareAndSwap(false, true) {
		return nil
	}

	ec := make(chan error, 1)
	go func() {
		defer close(ec)

		if err := rc.cli.Close(); err != nil {
			ec <- err
		}
	}()

	select {
	case err := <-ec:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func NewRedisClient(cfg RedisCfg) *RedisClient {
	if len(cfg.Addresses) == 0 {
		panic("addresses is required")
	}

	cluster := len(cfg.Addresses) > 1

	var cli redis.UniversalClient

	if cluster {
		cli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           cfg.Addresses,
			Username:        cfg.Username,
			Password:        cfg.Password,
			MaxIdleConns:    cfg.MaximumIdle,
			MinIdleConns:    cfg.MinimumIdle,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			PoolTimeout:     cfg.MaxWaitTimeout,
			ConnMaxIdleTime: cfg.MaxIdleTime,
			PoolSize:        cfg.Maximum,
		})
	} else {
		cli = redis.NewClient(&redis.Options{
			Addr:            cfg.Addresses[0],
			DB:              cfg.Database,
			Username:        cfg.Username,
			Password:        cfg.Password,
			MaxIdleConns:    cfg.MaximumIdle,
			MinIdleConns:    cfg.MinimumIdle,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			PoolTimeout:     cfg.MaxWaitTimeout,
			ConnMaxIdleTime: cfg.MaxIdleTime,
			PoolSize:        cfg.Maximum,
		})
	}

	return &RedisClient{
		cli: cli,
		cfg: cfg,
	}
}

type Action func(cli redis.UniversalClient) error

func (rc *RedisClient) With(a Action) error {
	return a(rc.cli)
}

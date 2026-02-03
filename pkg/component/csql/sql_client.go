package csql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"time"
)

const SqlLifetimeTag = "sql_client"

type SqlCfg struct {
	Schema      string
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	MaximumIdle int           // 最大闲置连接数
	Maximum     int           // 池中最大连接数
	MaxLifeTime time.Duration // 连接最长存活时间
	MaxIdleTime time.Duration // 最大闲置时间
	PingTimeout time.Duration // ping超时
}

type SqlClient struct {
	conn *dbr.Connection
	cfg  SqlCfg
}

func (sc *SqlClient) OnCreated(ec chan<- error) {
	ctx, cancel := context.WithTimeout(context.Background(), sc.cfg.PingTimeout)
	defer cancel()

	err := sc.conn.PingContext(ctx)
	if err != nil {
		ec <- err
	}

	lg := mylog.GetInitMarkerLogger()
	lg.Info().Msg("int sql client successfully")
}

func (sc *SqlClient) OnDispose(ctx context.Context) error {
	ec := make(chan error, 1)
	go func() {
		defer close(ec)

		if err := sc.conn.Close(); err != nil {
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

func NewSqlClient(cfg SqlCfg) (*SqlClient, error) {
	var dsn = buildDsn(cfg)

	if dsn == "" {
		panic("only mysql supported")
	}

	conn, err := dbr.Open(cfg.Schema, dsn, newSqlLogger())
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(cfg.Maximum)
	conn.SetConnMaxLifetime(cfg.MaxLifeTime)
	conn.SetMaxIdleConns(cfg.MaximumIdle)
	conn.SetConnMaxIdleTime(cfg.MaxIdleTime)

	sc := &SqlClient{
		conn: conn,
		cfg:  cfg,
	}

	return sc, nil
}

func buildDsn(cfg SqlCfg) string {
	if cfg.Schema == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)
	}

	return ""
}

type Action func(conn *dbr.Connection) error

func (sc *SqlClient) With(a Action) error {

	return a(sc.conn)
}

type SessAction func(sess *dbr.Session) error

func (sc *SqlClient) WithSess(sa SessAction) error {
	return sa(sc.conn.NewSession(nil))
}

type TransAction func(tx *dbr.Tx) error

func (sc *SqlClient) WithTrans(ta TransAction) error {
	se := sc.conn.NewSession(nil)

	tx, err := se.Begin()
	if err != nil {
		return err
	}

	defer tx.RollbackUnlessCommitted()

	if err = ta(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

type TransCtxAction func(ctx context.Context, tx *dbr.Tx) error

func (sc *SqlClient) WithTransCtx(ctx context.Context, tca TransCtxAction) error {
	return sc.WithTransAdvance(ctx, tca, nil)
}

func (sc *SqlClient) WithTransAdvance(ctx context.Context, tca TransCtxAction, ops *sql.TxOptions) error {
	se := sc.conn.NewSession(nil)

	tx, err := se.BeginTx(ctx, ops)
	if err != nil {
		return err
	}

	defer tx.RollbackUnlessCommitted()

	if err = tca(ctx, tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

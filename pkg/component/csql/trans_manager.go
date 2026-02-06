package csql

import (
	"context"
	"database/sql"
)

/*
业务层持有
*/
type TransManager interface {
	DoInTx(ctx context.Context, tca TransCtxAction) error

	DoInTxOptions(ctx context.Context, tca TransCtxAction, ops *sql.TxOptions) error
}

type transManager struct {
	sc *SqlClient
}

func NewTransManager(sc *SqlClient) TransManager {
	return &transManager{
		sc: sc,
	}
}

func (m *transManager) DoInTx(ctx context.Context, tca TransCtxAction) error {
	return m.sc.WithTransCtx(ctx, tca)

}

func (m *transManager) DoInTxOptions(ctx context.Context, tca TransCtxAction, ops *sql.TxOptions) error {
	return m.sc.WithTransAdvance(ctx, tca, ops)
}

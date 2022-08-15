package dao

import (
	"context"

	"git.bell.ai/Technology/go-delayer/internal/model"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/database/sql"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("config.txt").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("MySQL").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

func (d *dao) RawArticle(ctx context.Context, id int64) (art *model.Article, err error) {
	// get data from db
	return
}

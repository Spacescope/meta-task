package storage

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/mitchellh/mapstructure"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

type PostgreSQLParams struct {
	DSN     string
	MaxIdle int
	MaxOpen int
}

var _ Storage = (*PGSQL)(nil)

// PGSQL for postgre sql
type PGSQL struct {
	engine *xorm.Engine
}

func (p *PGSQL) Name() string {
	return "pgsql"
}

func (p *PGSQL) Sync(m ...interface{}) error {
	return p.engine.Sync2(m...)
}

// InitFromConfig init from config
func (p *PGSQL) InitFromConfig(ctx context.Context, storageCFG *config.Storage) error {
	var (
		err    error
		params PostgreSQLParams
	)

	if err = mapstructure.Decode(storageCFG.Params, &params); err != nil {
		return errors.Wrap(err, "mapstructure.Decode failed")
	}
	if params.DSN == "" {
		return errors.New("dsn can not empty")
	}

	p.engine, err = xorm.NewEngine("postgres", params.DSN)
	if err != nil {
		return errors.Wrap(err, "db init failed")
	}

	p.engine.Dialect().SetParams(map[string]string{"rowFormat": "DYNAMIC"})
	p.engine.SetSchema("")
	if params.MaxIdle > 0 {
		p.engine.SetMaxIdleConns(params.MaxIdle)
	}
	if params.MaxOpen > 0 {
		p.engine.SetMaxOpenConns(params.MaxOpen)
	}
	p.engine.SetMapper(names.GonicMapper{})

	p.engine.ShowSQL(false)

	return nil
}

// Existed judge model exist or not
func (p *PGSQL) Existed(m interface{}, height int64, version int) (bool, error) {
	count, err := p.engine.Where("height=? and version=?", height, version).Count(m)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Write insert one record into db

func (p *PGSQL) DelOldVersionAndWrite(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	session := p.engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	_, err := session.Where("height = ? and version < ?", height, version).Delete(t)
	if err != nil {
		return err
	}

	_, err = session.InsertOne(m)
	if err != nil {
		return err
	}
	session.Commit()

	return nil
}

func (p *PGSQL) DelOldVersionAndWriteMany(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	session := p.engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	_, err := session.Where("height = ? and version <= ?", height, version).Delete(t)
	if err != nil {
		return err
	}

	_, err = session.Insert(m)
	if err != nil {
		return err
	}
	session.Commit()

	return nil
}

package storage

import (
	"context"

	// "github.com/Spacescore/observatory-task/config"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
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

// Existed judge model exist or not
func (p *PGSQL) Existed(m interface{}, height int64, version int) (bool, error) {
	count, err := p.engine.Where("height=? and version=?", height, version).Count(m)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Write insert one record into db

func (p *PGSQL) Insert(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
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

func (p *PGSQL) Inserts(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
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

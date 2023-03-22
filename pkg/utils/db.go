package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"xorm.io/xorm/names"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"xorm.io/xorm"
)

const DBTYPE = "postgres"
const DBRetry = 3

const (
	TaskDB = "task_db"
)

var (
	EngineGroup map[string]*xorm.Engine
)

type EngineInfo struct {
	Key    string
	Schema string
	Tables []interface{}
}

func NewEngineGroup(ctx context.Context, ei *[]EngineInfo) map[string]*xorm.Engine {
	engineGroup := make(map[string]*xorm.Engine)

	for _, sei := range *ei {
		x, _ := InitDBEngine(ctx, sei.Schema, sei.Tables)
		engineGroup[sei.Key] = x
	}

	return engineGroup
}

// InitDBEngine In case of problems connecting to DB, retry connection.
func InitDBEngine(ctx context.Context, schema string, tables []interface{}) (x *xorm.Engine, err error) {
	log.Info("Beginning ORM engine initializations")

	for i := 0; i < DBRetry; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Aborted due to shutdown:\nin retry ORM engine initialization")
		default:
		}

		log.Infof("ORM engine initialization attempt #%d/%d...", i+1, DBRetry)

		if x, err = newEngine(ctx, schema); err == nil {
			break
		} else if i == DBRetry-1 {
			os.Exit(1)
			return nil, err
		}

		log.Errorf("ORM engine initialization attempt #%d/%d failed. Error: %v", i+1, DBRetry, err)
		time.Sleep(time.Second * 7)
	}

	if err = SyncTables(x, tables); err != nil {
		log.Info(err)
		return nil, fmt.Errorf("sync database struct error: %v", err)
	}

	return x, nil
}

func SyncTables(x *xorm.Engine, tables []interface{}) error {
	if tables == nil {
		return nil
	}
	return x.StoreEngine("InnoDB").Sync2(tables...)
}

func newEngine(ctx context.Context /*, migrateFunc func(*xorm.Engine) error*/, schema string) (x *xorm.Engine, err error) {
	if x, err = setEngine(schema); err != nil {
		return nil, err
	}

	x.ShowSQL(false)
	x.SetDefaultContext(ctx)
	x.SetMapper(names.GonicMapper{})

	if err = x.Ping(); err != nil {
		return nil, err
	}

	// if err = migrateFunc(x); err != nil {
	// 	return fmt.Errorf("migrate: %v", err)
	// }

	return x, nil
}

func setEngine(schema string) (*xorm.Engine, error) {
	x, err := sqlEngineInit(schema)
	if err != nil {
		log.Fatal("Init SQL engine error: \n", err)
	}

	x.SetMaxIdleConns(7)
	x.SetMaxOpenConns(30)
	// x.SetConnMaxLifetime()
	x.SetMapper(names.GonicMapper{})

	return x, nil
}

func sqlEngineInit(connStr string) (*xorm.Engine, error) {
	var engine *xorm.Engine
	var err error

	engine, err = xorm.NewEngine(DBTYPE, connStr)
	if err != nil {
		return nil, err
	}

	engine.Dialect().SetParams(map[string]string{"rowFormat": "DYNAMIC"})
	engine.SetSchema("")

	return engine, nil
}

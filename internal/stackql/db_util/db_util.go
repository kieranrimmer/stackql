package db_util

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/logging"
)

func GetDB(driverName string, dbName string, cfg dto.SQLBackendCfg) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("cannot init %s TCP connection with empty dsn", dbName)
	}
	db, err := sql.Open(driverName, dsn)
	retryCount := 0
	for {
		if retryCount >= cfg.InitMaxRetries || err == nil {
			break
		}
		time.Sleep(time.Duration(cfg.InitRetryInitialDelay) * time.Second)
		db, err = sql.Open(driverName, dsn)
		retryCount++
	}
	if err != nil {
		return nil, fmt.Errorf("%s db object setup error = '%s'", driverName, err.Error())
	}
	logging.GetLogger().Debugln(fmt.Sprintf("opened %s TCP db with connection string = '%s' and err  = '%v'", dbName, dsn, err))
	pingErr := db.Ping()
	retryCount = 0
	for {
		if retryCount >= cfg.InitMaxRetries || pingErr == nil {
			break
		}
		time.Sleep(time.Duration(cfg.InitRetryInitialDelay) * time.Second)
		pingErr = db.Ping()
		retryCount++
	}
	if pingErr != nil {
		return nil, fmt.Errorf("%s connection setup ping error = '%s'", dbName, pingErr.Error())
	}
	logging.GetLogger().Debugln(fmt.Sprintf("opened and pinged %s TCP db with connection string = '%s' and err  = '%v'", dbName, dsn, err))
	return db, nil
}

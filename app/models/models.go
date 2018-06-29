package models

import (

	// examples
	// https://upper.io/db.v3/examples
	"github.com/zhangpanyi/basebot/logger"

	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

// 数据库连接池
var pools *sqlbuilder.Database

// 连接数据库
func Connect(settings db.ConnectionURL) error {
	if pools != nil {
		return nil
	}

	db, err := sqlite.Open(settings)
	if err != nil {
		return err
	}

	db.SetLogging(false)

	if err = db.Ping(); err != nil {
		db.Close()
		logger.Errorf("Failed to connect to sqlite, %v", err)
	}
	pools = &db
	return nil
}

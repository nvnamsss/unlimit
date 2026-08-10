package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlAdapter struct {
	gormer      *gorm.DB
	isCommitted bool
}

// NewMySQLDatabase returns a new instance of DB.
func NewMySQLDatabase() GormDBAdapter {
	return &mysqlAdapter{}
}

// Open opens a DB connection.
func (db *mysqlAdapter) Open(connectionString string, config gorm.Config) error {
	gormDB, err := gorm.Open(mysql.Open(connectionString), &config)
	if err != nil {
		return err
	}

	db.gormer = gormDB
	return nil
}

// Begin starts a DB transaction.
func (db *mysqlAdapter) Begin() GormDBAdapter {
	tx := db.gormer.Begin()
	return &mysqlAdapter{
		gormer:      tx,
		isCommitted: false,
	}
}

// RollbackUselessCommitted rollbacks useless DB transaction committed.
func (db *mysqlAdapter) RollbackUselessCommitted() {
	if !db.isCommitted {
		db.gormer.Rollback()
	}
}

// Commit commits a DB transaction.
func (db *mysqlAdapter) Commit() {
	if !db.isCommitted {
		db.gormer.Commit()
		db.isCommitted = true
	}
}

// Close closes DB connection.
func (db *mysqlAdapter) Close() {
	sqlDB, err := db.gormer.DB()
	if err != nil {
		return
	}

	_ = sqlDB.Close()
}

// Gormer returns an instance of gorm.DB.
func (db *mysqlAdapter) Gormer() *gorm.DB {
	return db.gormer
}

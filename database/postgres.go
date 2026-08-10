package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postgresAdapter struct {
	gormer      *gorm.DB
	isCommitted bool
}

// NewDB returns a new instance of DB.
func NewPostgresDatabase() GormDBAdapter {
	return &postgresAdapter{}
}

// Open opens a DB connection.
func (db *postgresAdapter) Open(connectionString string, config gorm.Config) error {
	gormDB, err := gorm.Open(postgres.Open(connectionString), &config)
	if err != nil {
		return err
	}

	db.gormer = gormDB
	return nil
}

// Begin starts a DB transaction.
func (db *postgresAdapter) Begin() GormDBAdapter {
	tx := db.gormer.Begin()
	return &postgresAdapter{
		gormer:      tx,
		isCommitted: false,
	}
}

// RollbackUselessCommitted rollbacks useless DB transaction committed.
func (db *postgresAdapter) RollbackUselessCommitted() {
	if !db.isCommitted {
		db.gormer.Rollback()
	}
}

// Commit commits a DB transaction.
func (db *postgresAdapter) Commit() {
	if !db.isCommitted {
		db.gormer.Commit()
		db.isCommitted = true
	}
}

// Close closes DB connection.
func (db *postgresAdapter) Close() {
	sqlDB, err := db.gormer.DB()
	if err != nil {
		return
	}

	_ = sqlDB.Close()
}

// Gormer returns an instance of gorm.DB.
func (db *postgresAdapter) Gormer() *gorm.DB {
	return db.gormer
}

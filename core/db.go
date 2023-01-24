package core

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PasswordConnection defines the model that describes the table in which all the usernames
// and passwords that any peer attempts to access the system with will be stored.
type PasswordConnection struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"` // type:bigint for MySQL
	IPAddress string    `gorm:"index; type:mediumtext not null"`
	Username  string    `gorm:"index; not null"`
	Password  string    `gorm:"index; not null"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli"`
}

// KeyConnection defines the model that describes the table in which all the usernames
// and public key that any peer attempts to access the system with will be stored.
type KeyConnection struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"` // type:bigint for MySQL
	IPAddress string    `gorm:"index; type:mediumtext not null; unique_index:uidx_key_ip"`
	Username  string    `gorm:"index; not null"`
	Key       string    `gorm:"index; not null"`
	KeyHash   string    `gorm:"index; not null; unique_index:uidx_key_ip"`
	Type      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli"`
}

// ConnectDB connects to the database and returns the db object.
func ConnectDB(verbose bool) (*gorm.DB, error) {
	logLevel := logger.Silent

	if verbose {
		logLevel = logger.Info
	}

	// Connect to the database. Maybe this can change in the futurue to support more
	// types of databases (MySQL, MariaDB, PostgreSQL, etc).
	db, err := gorm.Open(sqlite.Open("honeyshell.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return nil, err
	}

	// Create all the tables and make sure all possible migrations are applied automatically.
	db.AutoMigrate(&PasswordConnection{})
	db.AutoMigrate(&KeyConnection{})

	return db, nil
}

package driver

import (
	"fmt"

	"github.com/jinzhu/gorm"

	_ "github.com/lib/pq" // initialize postgres driver
)

func OpenCn(host string, port string, user string, password string, dbName string, debug bool) (*gorm.DB, error) {
	cn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	db, err := gorm.Open("postgres", cn)
	if err != nil {
		return nil, fmt.Errorf("db connection failed: %w", err)
	}

	db.SingularTable(true)
	db.LogMode(debug)
	return db, nil
}

package e2e

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	mysqlGorm "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitializeDb() (*sql.DB, error) {
	v := LoadConfigFile()
	dbUrl := v.GetString("DB_URL")
	u, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "mysql" {
		dbUrl = fmt.Sprintf(("%s@tcp(%s)%s?%s"), u.User.String(), u.Host, u.Path, u.RawQuery)
	}
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		return nil, err
	}
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return nil, err
	}
	fmt.Println("Connected!")
	return db, nil
}

func InitializeGormDb() (*gorm.DB, error) {
	connectionString := "merico:merico@tcp(localhost:3306)/lake"
	db, err := gorm.Open(mysqlGorm.Open(connectionString))
	if err != nil {
		return nil, err
	}
	return db, nil
}

package database

import (
	"cw-cal/model"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "bornakh"
	password = "bornakh"
	dbName   = "cw_deadline"
)

type CwDatabase struct {
	*gorm.DB
}

func NewCwDatabase() (*CwDatabase, error) {
	pSqlConn := fmt.Sprintf("host= %s port=%d user = %s password = %s dbname=%s", host, port, user, password, dbName)
	db, err := gorm.Open(postgres.Open(pSqlConn), &gorm.Config{})
	return &CwDatabase{
		DB: db,
	}, err
}

func (db *CwDatabase) GetUser(ID int64) *model.User {
	return nil
}

func CheckError(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

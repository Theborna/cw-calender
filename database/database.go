package database

import (
	"cw-cal/model"
	"fmt"
	"log"

	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
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

func (db *CwDatabase) GetUser(sender *tele.User) (*model.User, error) {
	var gUser gormUser
	ID := sender.ID
	db.First(&gUser, ID)
	if gUser.ID == 0 {
		return nil, fmt.Errorf("user %d not found", ID)
	}
	return &Map([]gormUser{gUser}, func(gUser gormUser) model.User {
		var schedule []gormCal
		db.Find(&schedule, "UserID = ?", ID)
		return model.User{
			User: sender,
			Schedule: Map(schedule, func(gCal gormCal) model.Calendar {
				return model.Calendar{
					Link:   gCal.Link,
					Events: []model.Event{},
				}
			}),
		}
	})[0], nil
}

func (db *CwDatabase) AddUser(user model.User) error {
	return db.Model(gormUser{}).Create(&gormUser{
		Model: &gorm.Model{
			ID: uint(user.ID),
		},
	}).Error
}

func (db *CwDatabase) AddCal(cal *model.Calendar, ID int64) error {
	return db.Model(gormCal{}).Create(&gormCal{
		Model: &gorm.Model{
			ID: uint(uuid.New().ID()),
		},
		UserID: ID,
		Link:   cal.Link,
	}).Error
}

func CheckError(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

/*interfaces used to interact with db*/

/*User interface to save in database*/
type gormUser struct {
	*gorm.Model
}

/*
Calender interface to save in database,
only includes calendar link
*/
type gormCal struct {
	*gorm.Model
	UserID int64  `gorm:"us_id"`
	Link   string `gorm:"link"`
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

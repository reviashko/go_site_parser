package model

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // here
)

//ConnectionData struct
type ConnectionData struct {
	Host                   string
	Port                   int
	Dbname                 string
	User                   string
	Password               string
	MaxOpenConn            int
	MaxIdleConn            int
	MaxConnLifetimeMinutes int
}

//ToString function
func (c *ConnectionData) ToString() string {
	return fmt.Sprintf("host=%s port=%v dbname=%s user=%s password=%s sslmode=disable", c.Host, strconv.Itoa(c.Port), c.Dbname, c.User, c.Password)
}

//DB struct
type DB struct {
	*sqlx.DB
}

// InitDB function
func InitDB(connectionData ConnectionData) (*DB, error) {

	db, err := sqlx.Connect("postgres", connectionData.ToString())
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	//connection pool default settings
	db.SetMaxOpenConns(connectionData.MaxOpenConn)
	db.SetMaxIdleConns(connectionData.MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(connectionData.MaxConnLifetimeMinutes) * time.Minute)

	if err = db.Ping(); err != nil {
		log.Panic(err)
		return nil, err
	}

	return &DB{db}, nil
}

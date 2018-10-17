package hive

import (
	"time"

	"github.com/globalsign/mgo"
)

type system struct {
	db *mgo.Session
}

func NewSystem(MongoDBHosts []string, AuthDatabase, AuthUserName, AuthPassword string) (*system, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    MongoDBHosts,
		Timeout:  60 * time.Second,
		Database: AuthDatabase,
		Username: AuthUserName,
		Password: AuthPassword,
	}

	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	return &system{
		db: mongoSession,
	}, nil
}

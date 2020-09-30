package storage

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"go.etcd.io/bbolt"
	"path"
)

const DBPath = "db"

func GetDBPath() string {
	return path.Join(utils.GetSubFolder(DBPath), "data.db")
}

func GetDB() (*bbolt.DB, error) {
	return bbolt.Open(GetDBPath(), 0600, nil)
}

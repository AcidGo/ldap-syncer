package sources

import (
    "database/sql"
)

type MySQLConnector struct 

type MySQLSrc struct {
    syncMap         map[string]strings
    dbConn          *sql.DB
}

func (src *MySQLSrc) SetSyncMap(sm map[string]string) {
    src.syncMap = sm
}

func (src *MySQLSrc) Open(MySQLFlags) error {

}
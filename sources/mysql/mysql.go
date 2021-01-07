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

func (src *MySQLSrc) Open(i interface{}) error {
    f, ok := i.(MySQLFlags)
    if !ok {
        return errros.New("expecting src_mysql.MySQLFlags")
    }


}



func generateDsn(addr, db string) (string, error) {
    
}
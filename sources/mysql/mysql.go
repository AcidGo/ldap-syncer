package src_mysql

import (
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/AcidGo/ldap-syncer/lib"
    _ "github.com/go-sql-driver/mysql"
)

const (
    maxOpenConns    = 5
)

type MySQLSrc struct {
    syncMap         map[string]string
    dbConn          *sql.DB
    targetTable     string
}

func (src *MySQLSrc) SetSyncMap(sm map[string]string) {
    src.syncMap = sm
}

func (src *MySQLSrc) Open(i interface{}) error {
    f, ok := i.(MySQLFlags)
    if !ok {
        return errors.New("expecting src_mysql.MySQLFlags")
    }

    dsn, err := src.generateDsn(
        *f.Username,
        *f.Password,
        *f.ConnAddr,
        *f.TargetDB,
    )
    if err != nil {
        return err
    }

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }

    db.SetMaxOpenConns(maxOpenConns)

    if err = db.Ping(); err != nil {
        return err
    }

    src.dbConn = db

    tb := strings.TrimSpace(*f.TargetTable)
    if tb == "" {
        return fmt.Errorf("empty target table for MySQL: %s", *f.TargetTable)
    }

    src.targetTable = tb

    return nil
}

func (src *MySQLSrc) Close() {
    if src.dbConn != nil {
        src.dbConn.Close()
    }
}

func (src *MySQLSrc) Pull(pkField string) (*lib.EntryGroup, error) {
    fileds := pkField
    for _, val := range src.syncMap {
        if val == pkField {
            continue
        }
        fileds = fmt.Sprintf("%s, %s", fileds, val)
    }

    sql := fmt.Sprintf("select %s from %s", fileds, src.targetTable)
    rows, err := src.dbConn.Query(sql)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    cols, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    var res []map[string]string
    if rows.Next() {
        data := make(map[string]string)
        columns := make([]string, len(cols))
        columnPointers := make([]interface{}, len(cols))
        for i, _ := range columns {
            columnPointers[i] = &columns[i]
        }

        rows.Scan(columnPointers...)

        for i, colName := range cols {
            data[colName] = columns[i]
        }

        res = append(res, data)
    }

    eg, err := lib.MapSliceToGroup(pkField, res)
    if err != nil {
        return nil, err
    }

    return eg, nil
}

func (src *MySQLSrc) generateDsn(user, passwd, addr, db string) (string, error) {
    if len(strings.Split(addr, ":")) != 2 {
        return "", fmt.Errorf("invalid mysql connect addr: %s", addr)
    }

    return fmt.Sprintf("%s:%s@tcp(%s)/%s", user, passwd, addr, db), nil
}
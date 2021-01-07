package ldap

import (
    "errors"
    "fmt"
    "log"
    "strings"

    "github.com/AcidGo/ldap-syncer/lib"
    ldaplib "github.com/go-ldap/ldap/v3"
)

type OpUpdate []*ldaplib.ModifyRequest
type OpDelete []*ldaplib.DelRequest
type OpInsert []*ldaplib.AddRequest

type LdapDst struct {
    conn        *ldaplib.Conn
    workDn      string
    syncMap     map[string]string
    pkFiled     string
    opUpdate    OpUpdate
    opDelete    OpDelete
    opInsert    OpInsert
}

func NewLdapDst(ldapAddr, bindUser, bindPasswd, workDn string) (*LdapDst, error) {
    conn, err := ldaplib.Dial("tcp", ldapAddr)
    if err != nil {
        return nil, err
    }

    err = conn.Bind(bindUser, bindPasswd)
    if err != nil {
        return nil, err
    }

    testRequest := ldaplib.NewSearchRequest(
        workDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(objectclass=*)",
        []string{},
        nil,
    )

    _, err = conn.Search(testRequest)
    if err != nil {
        return nil, err
    }

    return &LdapDst{
        conn: conn,
        workDn: workDn,
    }, nil
}

func (l *LdapDst) Close() {
    if l.conn != nil {
        l.conn.Close()
    }
}

func (l *LdapDst) SetSyncMap(sm map[string]string) {
    l.syncMap = sm
}

func (l *LdapDst) GetSyncMap() map[string]string {
    return l.syncMap
}

func (l *LdapDst) Parse(pkFiled string, srcGroup *lib.EntryGroup) error {
    var err error
    var dnPrefix string

    searchRequest := ldaplib.NewSearchRequest(
        l.workDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(objectclass=*)",
        []string{},
        nil,
    )

    sr, err := l.conn.Search(searchRequest)
    if err != nil {
        return err
    }

    lGroup, err := lib.NewEntryGroup(pkFiled)
    if err != nil {
        return err
    }

    for _, e := range sr.Entries {
        lRow, err := lib.LdapEntryToRow(pkFiled, l.syncMap, e)
        if err == lib.PrimaryFieldNotFound {
            log.Printf("ignore dn [%s] because of empty primary key\n", e.DN)
            continue
        } else if err != nil {
             return err
        }

        if dnPrefix == "" && strings.Index(e.DN, l.workDn) != -1 {
            dnPrefix = strings.Split(strings.Split(e.DN, l.workDn)[0], "=")[0]
        }

        lRow.SetDN(e.DN)
        err = lGroup.AddRow(lRow)
        if err != nil {
            return err
        }
    }

    insert, update, delete, err := lib.EntryGroupDiff(srcGroup, lGroup)
    if err != nil {
        return err
    }
    log.Printf("after LDAP entry group diff, get length of insert is: %d\n", len(insert))
    log.Printf("after LDAP entry group diff, get length of update is: %d\n", len(update))
    log.Printf("after LDAP entry group diff, get length of delete is: %d\n", len(delete))

    updateUeqList, err := generateOpUpdate(update)
    if err != nil {
        return err
    }
    l.opUpdate = updateUeqList

    deleteReqList, err := generateOpDelete(delete)
    if err != nil {
        return err
    }
    l.opDelete = deleteReqList

    insertReqList, err := generateOpInsert(l.workDn, dnPrefix, insert)
    if err != nil {
        return err
    }
    l.opInsert = insertReqList

    return nil
}

func (l *LdapDst) Sync() error {
    var err error

    // working for insert operation
    for _, req := range l.opInsert {
        err = l.conn.Add(req)
        if err != nil {
            return fmt.Errorf("get an error when insert %s: %v", req.DN, err)
        }
    }

    // working for update operation
    for _, req := range l.opUpdate {
        err = l.conn.Modify(req)
        if err != nil {
            return fmt.Errorf("get an error when modify %s: %v", req.DN, err)
        }
    }

    // working for delete operation
    for _, req := range l.opDelete {
        err = l.conn.Del(req)
        if err != nil {
            return fmt.Errorf("get an error when delete %s: %v", req.DN, err)
        }
    }

    return err
}

func (l *LdapDst) ParsePrint() error {
    log.Println("----------> ParsePrint <----------")

    // print for insert operation
    log.Printf("########## insert operation: %d\n", len(l.opInsert))
    for _, req := range l.opInsert {
        log.Printf("DN: %s\n", req.DN)
        log.Println("OP:")
        for _, a := range req.Attributes {
            log.Printf("\t%s: %v\n", a.Type, a.Vals)
        }
    }
    log.Println("########## EOF insert operation")

    // print for update operation
    log.Printf("########## update operation: %d\n", len(l.opUpdate))
    for _, req := range l.opUpdate {
        log.Printf("DN: %s\n", req.DN)
        log.Println("OP:")
        for _, c := range req.Changes {
            log.Printf("\t%v\n", c.Modification.Vals)
        }
    }
    log.Println("########## EOF update operation")

    // print for delete operation
    log.Printf("########## delete operation: %d\n", len(l.opDelete))
    for _, req := range l.opDelete {
        log.Printf("DN: %s\n", req.DN)
    }
    log.Println("########## EOF delete operation")

    log.Println("--------> EOF ParsePrint <--------")

    return nil
}

func generateOpInsert(dn string, rows []*lib.EntryRow) ([]*ldaplib.AddRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    for idx, e := range rows {
        if dn == "" {
            return []*ldaplib.AddRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewAddRequest(dn, nil)
        for k, val := range e.GetRow() {
            req.Attribute(k, val)
        }
        reqList[idx] = req
    }

    return reqList, err
}

func generateOpUpdate(rows []*lib.EntryRow) ([]*ldaplib.ModifyRequest, error) {
    var err error
    reqList := make([]*ldaplib.ModifyRequest, len(rows))

    for idx, e := range rows {
        dn := e.GetDN()
        if dn == "" {
            return []*ldaplib.ModifyRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewModifyRequest(dn, nil)
        for k, val := range e.GetRow() {
            req.Replace(k, val)
        }
        reqList[idx] = req
    }

    return reqList, err
}

func generateOpDelete(rows []*lib.EntryRow) ([]*ldaplib.DelRequest, error) {
    var err error
    reqList := make([]*ldaplib.DelRequest, len(rows))

    for idx, e := range rows {
        dn := e.GetDN()
        if dn == "" {
            return []*ldaplib.DelRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewDelRequest(dn, nil)
        reqList[idx] = req
    }

    return reqList, err
}
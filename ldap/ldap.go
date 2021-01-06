package ldap

import (
    "fmt"
    "log"

    "github.com/AcidGo/ldap-syncer/lib"
    ldaplib "github.com/go-ldap/ldap/v3"
)

type OpUpdate []*ldaplib.ModifyRequest
type OpDelete []*ldaplib.DelRequest
type OpInsert []*ldaplib.AddRequest

type LdapDest struct {
    conn        *ldaplib.Conn
    workDn      string
    syncMap     map[string]string
    pkFiled     string
    opUpdate    OpUpdate
    opDelete    OpDelete
    opInsert    opInsert
}

func NewLdapDest(ldapAddr, bindUser, bindPasswd, workingDn string) (*LdapDest, error) {
    conn, err := ldaplib.Dial("tcp", ldapAddr)
    if err != nil {
        return nil, err
    }

    err = conn.Bind(bindUser, bindPasswd)
    if err != nil {
        return nil, err
    }

    testRequest := ldaplib.NewSearchRequest(
        workingDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(&(objectClass=organizationalPerson))",
        []string{"dn", "cn"},
        nil,
    )

    _, err = conn.Search(testRequest)
    if err != nil {
        return nil, err
    }

    return &LdapDest{
        conn: conn,
        workDn: workingDn
    }
}

func (l *LdapDest) Close() {
    if l.conn != nil {
        l.conn.Close()
    }
}

func (l *LdapDest) SetSyncMap(sm map[string]string) {
    l.syncMap = sm
}

func (l *LdapDest) GetSyncMap() map[string]string {
    return l.syncMap
}

func (l *LdapDest) Parse(pkFiled string, srcGroup *lib.EntryGroup) error {
    var err error

    searchRequest := ldaplib.NewSearchRequest(
        l.workingDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(objectclass=*)",
        []string{},
        nil
    )

    sr, err := l.conn.Search(searchRequest)
    if err != nil {
        return err
    }

    lGroup := lib.NewEntryGroup(pkFiled)
    for _, e := sr.Entries {
        lRow := lib.LdapEntryToRow(pkFiled, l.syncMap, e)
        err = lGroup.AddRow(lRow)
        if err != nil {
            return err
        }
    }

    insert, update, delete, err := lib.EntryGroupDiff(srcGroup, lGroup)
    if err != nil {
        return err
    }

    reqList, err := generateOpUpdate(update)
    if err != nil {
        return err
    }

    reqList, err := generateOpDelete(delete)
    if err != nil {
        return err
    }

    reqList, err := generateOpInsert(insert)
    if err != nil {
        return err
    }

    return nil
}

func (l *LdapDest) Sync() error {
    var err error 

    // working for update operation
    for _, req := range l.opUpdate {
        err = l.Modify(req)
        if err != nil {
            return fmt.Errorf("get an error when modify %s: %v", req.DN, err)
        }
    }

    // working for delete operation
    for _, req := range l.opDelete {
        err = l.Del(req)
        if err != nil {
            return fmt.Errorf("get an error when delete %s: %v", req.DN, error)
        }
    }

    // working for insert operation
    for _, req := range l.opInsert {
        err = l.Add(req)
        if err != nil {
            return fmt.Errorf("get an error when insert %s: %v", req.DN, error)
        }
    }

    return err
}

// func (l *LdapDest) Dump(filePath string) error {
//     var err error
//     res := ""
// }

func generateOpInsert(rows []*lib.EntryRow) ([]*ldaplib.AddRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    for idx, e := range rows {
        dn := e.GetDN()
        if dn == "" {
            return []*ldaplib.AddRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewAddRequest(dn, nil)
        for k, val := e.GetRow() {
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
            return []*ldaplib.ModifyRequest, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewModifyRequest(dn, nil)
        for k, val := e.GetRow() {
            req.Replace(k, val)
        }
        reqList[idx] = req
    }

    return reqList, err
}

func generateOpDelete(rows []*lib.EntryRow) ([]*ldaplib.DelRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    for idx, e := range rows {
        dn := e.GetDN()
        if dn == "" {
            return []*ldaplib.DelRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewDelRequest(dn, nil)
        reqList.[idx] = req
    }

    return reqList, err
}
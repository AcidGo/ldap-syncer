package ldap

import (
    "fmt"
    "log"

    "github.com/AcidGo/ldap-syncer/lib"
    ldaplib "github.com/go-ldap/ldap/v3"
)

type Op map[*ldaplib.Entry]*ldaplib.AddRequest
type OpUpdate Op
type OpDelete Op
type OpInsert Op

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


}

func GenerateOpInsert(dn string, rows []*lib.EntryRow) ([]*ldaplib.AddRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    for idx, e := range rows {
        req := ldaplib.NewAddRequest(dn, nil)
        for k, val := e.GetRow() {
            req.Attribute(k, val)
        }
        reqList[idx] = req
    }

    return reqList, err
}

func GenerateOpUpdate(dn string, rows []*lib.EntryRow) ([]*ldaplib.ModifyRequest, error) {
    
}

func GenerateOpDelete(dn string, rows []*lib.EntryRow) ([]*ldaplib.DelRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    
}
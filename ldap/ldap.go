package ldap

import (
    "fmt"
    "log"

    ldaplib "github.com/go-ldap/ldap/v3"
)

type Op map[*ldaplib.Entry]*ldaplib.AddRequest
type OpUpdate Op
type OpDelete Op
type OpInsert Op

type EntryRow map[string][]string

type LdapDest struct {
    conn        *ldaplib.Conn
    workDn      string
    syncMap     map[string]string
    opUpdate    OpUpdate
    opDelete    OpDelete
    opInsert    opInsert
}

func NewLdapDest(ldapAddr, bindUser, bindPasswd, workingDn string) (*LdapDest, error) {
    conn, err := ldaplib.Dial(ldapAddr)
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

    sr, err := conn.Search(testRequest)
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

func (l *LdapDest) Parse(sourceSet map[string]EntryRow) error {
    dnData := make([]map[string][]string)
    needInsertSet := make([]map[string][]string)
    for k, v := range sourceSet {
        needInsertSet[k] = v
    }

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

    for _, entry := range sr.Entries {
        attrMap := make(map[string][]string)
        for _, attr := range entry.Attributes {
            attrMap[attr.Name] = attrMap.Values
        }
        dnData = append(dnData, attrMap)
    }


}

func (l *LdapDest) newAddRequest()
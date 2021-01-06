package ldap

import (
    "fmt"
    "log"

    "github.com/AcidGo/ldap-syncer/sources"
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

func (l *LdapDest) Parse(sourceSet *SourceSetter) error {
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

    for _, e := range sr.Entries {
        eDn := e.DN
        eAttrRow := source.NewEntryRow()
        for _, a := range e.Attributes {
            eAttrRow.AddValues(a.Name, a.Values)
        }
        if val, ok := eAttrRow.GetValues(sourceSet.PrimaryKey()); ok {
            if 
        } else {
            return fmt.Errorf("not found SourceSetter primary key %v for lookup", sourceSet.PrimaryKey())
        }
    }
}


func (l *LdapDest) newAddRequest()

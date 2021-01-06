package ldap

import (
    "testing"
)

func getLdapDest() *LdapDest {
    l, err := NewLdapDest(
        "192.168.66.131:389",
        "cn=admin,dc=acidgo,dc=com",
        "suan",
        "ou=Users,dc=acidgo,dc=com",
    )
    if err != nil {
        t.Fatal("get an err when get LdapDest:", err)
    }
    return l
}

func TestLdapConn(t *testing.T) {
    l := getLdapDest()
    defer l.Close()
}
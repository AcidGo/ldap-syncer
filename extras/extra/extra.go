package extras

import (
    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/sources/source"
)

type Extrar interface {
    BindSource(sources.Sourcer) error
    BindLdap(*ldap.LdapDst) error
    Parse(interface{}) error
    ParsePrint()
    Run() error
}

type Flags struct {}
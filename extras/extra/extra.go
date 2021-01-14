package extras

import (
    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/sources/source"
)

type Extrar interface {
    BindSource(sources.sourcer)
    BindLDAP(*ldap.LdapDst)
    Parse() error
    ParsePrint()
    Run() error
}
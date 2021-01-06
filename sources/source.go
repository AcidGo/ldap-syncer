package sources

import (
    "fmt"
    "errors"

    "github.com/AcidGo/ldap-syncer/lib"
    "github.com/AcidGo/ldap-syncer/utils"
)

type Sourcer interface {
    SetSyncMap(map[string]string)
    Open() error
    Pull() *lib.EntryGroup
}
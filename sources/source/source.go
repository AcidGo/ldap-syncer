package sources

import (
    "github.com/AcidGo/ldap-syncer/lib"
)

type Sourcer interface {
    SetSyncMap(map[string]string)
    Open(interface{}) error
    Close()
    Pull(string) (*lib.EntryGroup, error)
}

type Flags struct {}
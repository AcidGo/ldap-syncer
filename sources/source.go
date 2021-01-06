package sources

import (
    "fmt"
    "errors"

    "github.com/AcidGo/ldap-syncer/lib"
)

type Sourcer interface {
    SetSyncMap(map[string]string)
    Open(Flags) error
    Close()
    Pull(string) (*lib.EntryGroup, error)
}

type Flags struct {}
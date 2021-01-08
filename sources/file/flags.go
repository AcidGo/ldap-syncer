package src_file

import (
    "github.com/AcidGo/ldap-syncer/sources/source"
)

type FileFlags struct {
    sources.Flags
    Path        *string
}
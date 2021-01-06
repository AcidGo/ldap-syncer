package sources

import (
    "fmt"
    "errors"

    "github.com/AcidGo/ldap-syncer/utils"
)

type SourceSetter interface {
    AddEntryRow(string, EntryRow)
    GetEntryRow(string) (EntryRow, bool)
    PrimaryKey() string
    ResetState()
    ParseEntryRow(EntryRow) (bool, EntryRow)
    GetMissEntryRow() []EntryRow
}

type EntryRow map[string][]string

type SourceSet struct {
    set             map[string]EntryRow
    primaryKey      string
    missEntryRow    map[string]EntryRow
}

type Sourcer interface {
    GenerateSrcSet() (*SourceSet, error)
}

func NewEntryRow() *EntryRow {
    return &EntryRow{}
}

func (e *EntryRow) AddValues(k string, v []string) {
    e[k] = v
}

func (e *EntryRow) GetValues(k string) ([]string, bool) {
    return e[k]
}

func NewSourceSet(primaryKey string) (*SourceSet, error) {
    if primaryKey == "" {
        return nil, errors.New("primary key for generate SourceSet is empty")
    }
    return &SourceSet{
        set: make(map[string]EntryRow),
        primaryKey: primaryKey,
        missEntryRow: make(map[string]EntryRow)
    }, nil
}

func (ss *SourceSet) AddEntryRow(k string, er EntryRow) {
    ss.set[k] = er
}

func (ss *SourceSet) GetEntryRow(k string) (EntryRow, bool) {
    res, ok := set[k]
    return res, ok
}

func (ss *SourceSet) PrimaryKey() string {
    return ss.primaryKey
}

func (ss *SourceSet) ResetState() {
    ss.missEntryRow := make(map[string]EntryRow)
}

func (ss *SourceSet) GetMissEntryRow() map[string]EntryRow {
    return ss.missEntryRow
}

func (ss *SourceSet) ParseEntryRow(pk string, e EntryRow) (bool, EntryRow, error) {
    sER, sOk := ss.set[pk]
    eVal, eOk := e.GetValues(pk)
    if eOk && !sOk {
        // need to delete
        return true, nil, nil
    } 
    if !eOk && sOk {
        // need to insert
        return true, sER, nil
    }
    if utils.IsSameStringList(eVal, sVal) {
        // nothing to do
        return false, nil, nil
    } else {
        // need to update
        return false, 
    }


}
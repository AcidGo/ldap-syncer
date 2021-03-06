package lib

import (
    "errors"
    "fmt"

    "github.com/AcidGo/ldap-syncer/utils"
    ldaplib "github.com/go-ldap/ldap/v3"
)

var (
    PrimaryFieldNotFound        = errors.New("primary key not found")
)

type EntryRow struct {
    pkField         string
    pkName          string
    dn              string
    data            map[string][]string
}

func NewEntryRow(pkField string, pkName string) (*EntryRow, error) {
    if pkField == "" || pkName == "" {
        return nil, errors.New("the field of primary key or name of primary key is empyt")
    }
    return &EntryRow{
        pkField: pkField, 
        pkName: pkName,
        data: make(map[string][]string),
    }, nil
}

func (e *EntryRow) PKField() string {
    return e.pkField
}

func (e *EntryRow) PKName() string {
    return e.pkName
}

func (e *EntryRow) SetValue(k string, v []string) {
    e.data[k] = v
}

func (e *EntryRow) SetDN(dn string) {
    e.dn = dn
}

func (e *EntryRow) GetValue(k string) ([]string, bool) {
    val, ok := e.data[k]
    return val, ok
}

func (e *EntryRow) GetDN() string {
    return e.dn
}

func (e *EntryRow) GetRow() map[string][]string {
    return e.data
}

func (e *EntryRow) Reflect(syncMap map[string]string) {
    for k, v := range e.data {
        if mVal, ok := syncMap[k]; ok {
            delete(e.data, k)
            e.data[mVal] = v
        }
    }
}

func (e *EntryRow) IsSame(d *EntryRow) bool {
    if e.pkField != d.PKField() || e.pkName != d.PKName() {
        return false
    }

    dData := d.GetRow()
    if len(e.data) != len(dData) {
        return false
    }
    for key, val := range e.data {
        if dVal, ok := dData[key]; !ok {
            return false
        } else {
            if !utils.IsSameStringList(val, dVal) {
                return false
            }
        }
    }
    return true
}

type EntryGroup struct {
    pkField         string
    set             map[string]*EntryRow
}

func NewEntryGroup(pkField string) (*EntryGroup, error) {
    if pkField == "" {
        return nil, errors.New("the field of primary key is emtpy")
    }
    return &EntryGroup{
        pkField: pkField,
        set: make(map[string]*EntryRow),
    }, nil
}

func (eg *EntryGroup) PKField() string {
    return eg.pkField
}

func (eg *EntryGroup) AddRow(e *EntryRow) error {
    var err error

    if e.PKField() != eg.pkField {
        return errors.New("the row's primary key field is not equal for group")
    }
    eg.set[e.PKName()] = e

    return err
}

func (eg *EntryGroup) GetRow(k string) (*EntryRow, bool) {
    val, ok := eg.set[k]
    return val, ok
}

func (eg *EntryGroup) GetGroup() map[string]*EntryRow {
    return eg.set
}

func LdapEntryToRow(pkField string, syncMap map[string]string, e *ldaplib.Entry) (*EntryRow, error) {
    if pkField == "" {
        return nil, errors.New("the field of primary key filed is emtpy")
    }

    // for _, a := range e.Attributes {
    //     fmt.Println(a.Name)
    //     fmt.Println(a.Values)
    // }

    pkName := e.GetAttributeValue(pkField)
    if pkName == "" {
        return nil, PrimaryFieldNotFound
    }
    row, err := NewEntryRow(pkField, pkName)
    if err != nil {
        return nil, err
    }

    for _, dstAttrName := range syncMap {
        row.SetValue(
            dstAttrName,
            e.GetAttributeValues(dstAttrName),
        )
    }

    return row, nil
}

func MapSliceToGroup(pkField string, ms []map[string]string) (*EntryGroup, error) {
    if pkField == "" {
        return nil, errors.New("the field of primary key filed is emtpy")
    }

    eg, err := NewEntryGroup(pkField)
    if err != nil {
        return nil, err
    }

    for _, m := range ms {
        pkName, ok := m[pkField]
        if !ok {
            return nil, fmt.Errorf("not found the primary key field: %s", pkField)
        }

        er, err := NewEntryRow(pkField, pkName)
        if err != nil {
            return nil ,err
        }

        for key, val := range m {
            if key == pkField {
                continue
            }

            er.SetValue(
                key,
                []string{val},
            )
        }

        eg.AddRow(er)
    }

    return eg, nil
}

func EntryGroupDiff(sm map[string]string, src, dst *EntryGroup) (insert, update, delete []*EntryRow, err error) {
    for srcName, srcRow := range src.GetGroup() {
        if dstRow, ok := dst.GetRow(srcName); !ok {
            srcRow.Reflect(sm)
            insert = append(insert, srcRow)
        } else {
            if !srcRow.IsSame(dstRow) {
                srcRow.Reflect(sm)
                update = append(update, srcRow)
            }
        }
    }

    for dstName, dstRow := range dst.GetGroup() {
        if _, ok := src.GetRow(dstName); !ok {
            delete = append(delete, dstRow)
        }
    }

    return 
}
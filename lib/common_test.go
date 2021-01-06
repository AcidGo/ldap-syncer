package lib

import (
    "strconv"
    "testing"

    "github.com/AcidGo/ldap-syncer/utils"
)

func getEntryRow(seq int, change bool) *EntryRow {
    var pkField string
    var pkName string

    if change {
        pkField = "field" + strconv.Itoa(seq)
    } else {
        pkField = "field"
    }
    pkName = "name" + strconv.Itoa(seq)

    er, _ := NewEntryRow(pkField, pkName)
    for i := 0; i <= seq; i++ {
        k := strconv.Itoa(seq)
        v := make([]string, 0)
        for j := 0; j <= i; j++ {
            v = append(v, strconv.Itoa(j))
        }
        er.SetValue(k, v)
    }

    return er
}

func TestEntryRow(t *testing.T) {
    var err error

    const (
        pkField = "testing-pkfield"
        pkName = "testing-pkname"
    )

    er, err := NewEntryRow(pkField, pkName)
    if err != nil {
        t.Errorf("new an EntryRow is failed: %v", err)
    }

    data := map[string][]string {
        "test1": []string{"a", "b"},
        "test2": []string{"b"},
    }

    for k, v := range data {
        er.SetValue(k, v)
    }
    for k, v := range data {
        if val, ok := er.GetValue(k); !ok || !utils.IsSameStringList(v, val) {
            t.Errorf("EntryRow SetValue/GetValue is not expected same for %s", k)
        }
    }
    if _, ok := er.GetValue("__NULL__"); ok {
        t.Errorf("EntryRow GetValue is not expected for getting an null value")
    }

    dn := "DN"
    er.SetDN(dn)
    if dn != er.GetDN() {
        t.Errorf("EntryRow SetDN/GetDN is not expected same for %s", dn)
    }

    if val := er.GetRow(); len(val) != len(data) {
        t.Errorf("EntryRow GetRow is not expected for getting input data")
    }

    er1, _ := NewEntryRow(pkField, pkName)
    for k, v := range data {
        er1.SetValue(k, v)
    }
    er2, _ := NewEntryRow(pkField+"__", pkName)
    for k, v := range data {
        er2.SetValue(k, v)
    }
    er3, _ := NewEntryRow(pkField, pkName+"___")
    for k, v := range data {
        er3.SetValue(k, v)
    }
    if !er.IsSame(er1) {
        t.Errorf("EntryRow IsSame for er1 expected true but false")
    }
    if er.IsSame(er2) {
        t.Errorf("EntryRow IsSame for er2 expected false but true")
    }
    if er.IsSame(er3) {
        t.Errorf("EntryRow IsSame for er3 expected flase but true")
    }
}

func TestEntryGroupDiff(t *testing.T) {
    var err error

    eg1, err := NewEntryGroup("field")
    if err != nil {
        t.Errorf("get an error when new EntryGroup: %v", err)
    }
    eg2, _ := NewEntryGroup("field")

    er1_1 := getEntryRow(1, false)
    er1_2 := getEntryRow(2, false)
    er1_3 := getEntryRow(3, false)

    er2_1 := getEntryRow(1, false)
    er2_2 := getEntryRow(2, false)
    er2_3 := getEntryRow(3, false)

    eg1.AddRow(er1_1)
    eg2.AddRow(er2_1)
    i, u, d, err := EntryGroupDiff(eg1, eg2)
    if len(i) != 0 || len(u) != 0 || len(d) != 0 || err != nil {
        t.Errorf("EntryGroupDiff test-1 failed")
    }

    eg2.AddRow(er2_2)
    i, u, d, err = EntryGroupDiff(eg1, eg2)
    if len(i) != 0 || len(u) != 0 || len(d) != 1 || err != nil {
        t.Errorf("EntryGroupDiff test-2 failed, i: %d, u: %d, d: %d", len(i), len(u), len(d))
    }

    eg1.AddRow(er1_2)
    eg1.AddRow(er1_3)
    i, u, d, err = EntryGroupDiff(eg1, eg2)
    if len(i) != 1 || len(u) != 0 || len(d) != 0 || err != nil {
        t.Errorf("EntryGroupDiff test-3 failed, i: %d, u: %d, d: %d", len(i), len(u), len(d))
    }

    eg2.AddRow(er2_3)
    i, u, d, err = EntryGroupDiff(eg1, eg2)
    if len(i) != 0 || len(u) != 0 || len(d) != 0 || err != nil {
        t.Errorf("EntryGroupDiff test-4 failed, i: %d, u: %d, d: %d", len(i), len(u), len(d))
    }
}
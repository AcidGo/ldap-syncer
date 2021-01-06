package src_file

import (
    "errors"
    "io/ioutil"
    "os"
    "strings"

    "github.com/AcidGo/ldap-syncer/lib"
)

type FileSrc struct {
    filePath        string
    fn              *os.File
    syncMap         map[string]string
}

func (src *FileSrc) SetSyncMap(sm map[string]string) {
    src.syncMap = sm
}

func (src *FileSrc) Open(i interface{}) error {
    f, ok := i.(FileFlags)
    if !ok {
        return errors.New("expecting src_file.FileFLags")
    }
    if _, err := os.Stat(*f.Path); os.IsNotExist(err) {
        return err
    }

    src.filePath = *f.Path
    fn, err := os.Open(src.filePath)
    if err != nil {
        return err
    }

    src.fn = fn

    return nil
}

func (src *FileSrc) Close() {
    if src.fn != nil {
        src.fn.Close()
    }
}

func (src *FileSrc) Pull(pkField string) (*lib.EntryGroup, error) {
    b, err := ioutil.ReadAll(src.fn)
    if err != nil {
        return nil, err
    }

    eg, err := lib.NewEntryGroup(pkField)
    if err != nil {
        return nil, err
    }

    s := string(b)
    for _, line := range strings.Split(s, "\n") {
        var _pkField string
        var _pkName string
        var er *lib.EntryRow
        var err error

        for idx, chunk := range strings.Split(line, "|") {
            if idx == 0 {
                _pkField = strings.Split(chunk, ":")[0]
                _pkName = strings.Split(chunk, ":")[1]
                er, err = lib.NewEntryRow(_pkField, _pkName)
                if err != nil {
                    return nil, err
                }
            } else {
                k := strings.Split(chunk, ":")[0]
                v := strings.Split(chunk, ":")[1]
                er.SetValue(k, []string{v})
            }
        }

        eg.AddRow(er)
    }

    return eg, nil
}
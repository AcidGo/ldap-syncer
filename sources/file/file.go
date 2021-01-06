package sources

import (
    "os"

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

func (src *FileSrc) Open(f FileFlags) error {
    if _, err := os.Stat(*f.Path); os.IsNotExists(err) {
        return err
    }

    src.filePath = *f.Path
    fn, err := os.Open(src.filePath)
    if err != nil {
        return err
    }

    src.fn = fn
}

func (src *FileSrc) Close() {
    if src.fn != nil {
        src.fn.Close()
    }
}

func (src *FileSrc) Pull(pkField string) (*lib.EntryGroup, error) {
    b, err := ioutil.ReadFile(src.fn)
    if err != nil {
        return nil, err
    }

    eg, err := lib.NewEntryGroup(pkField)
    if err != nil {
        return err
    }

    s := string(b)
    for _, line := range strings.Split(s, "\n") {
        _pkField := ""
        _pkName := ""
        attr := make(map[string][]string)
        for idx, chunk := range strings.Split(line, "|") {
            if idx == 0 {
                _pkField = strings.Split(chunk, ":")[0]
                _pkName = strings.Split(chunk, ":")[1]
            } else {
                k := strings.Split(chunk, ":")[0]
                v := strings.Split(chunk, ":")[1]
                attr[k] = []string{v}
            }
        }
        er, err := lib.NewEntryRow(_pkField, _pkName)
        if err != nil {
            return nil, err
        }
        
    }
}
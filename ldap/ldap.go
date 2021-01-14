package ldap

import (
    "errors"
    "fmt"
    "log"
    "strings"

    "github.com/AcidGo/ldap-syncer/lib"
    "github.com/AcidGo/ldap-syncer/utils"
    ldaplib "github.com/go-ldap/ldap/v3"
)

var passwdFieldSample = []string {
    "userpassword",
    "userpasswd",
    "passwd",
    "password",
}

type OpUpdate []*ldaplib.ModifyRequest
type OpDelete []*ldaplib.DelRequest
type OpInsert []*ldaplib.AddRequest

type LdapDst struct {
    conn            *ldaplib.Conn
    workDn          string
    syncMap         map[string]string
    pkFiled         string
    usedobjectClass []string
    encryptType     string
    hash            cryptFunc
    opUpdate        OpUpdate
    opDelete        OpDelete
    opInsert        OpInsert
}

func NewLdapDst(ldapAddr, bindUser, bindPasswd, workDn string) (*LdapDst, error) {
    conn, err := ldaplib.Dial("tcp", ldapAddr)
    if err != nil {
        return nil, err
    }

    err = conn.Bind(bindUser, bindPasswd)
    if err != nil {
        return nil, err
    }

    testRequest := ldaplib.NewSearchRequest(
        workDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(objectclass=*)",
        []string{},
        nil,
    )

    _, err = conn.Search(testRequest)
    if err != nil {
        return nil, err
    }

    return &LdapDst{
        conn: conn,
        workDn: workDn,
    }, nil
}

func (l *LdapDst) Close() {
    if l.conn != nil {
        l.conn.Close()
    }
}

func (l *LdapDst) SetSyncMap(sm map[string]string) {
    l.syncMap = sm
}

func (l *LdapDst) SetUsedObjectClass(oc []string) {
    l.usedobjectClass = oc
}

func (l *LdapDst) GetSyncMap() map[string]string {
    return l.syncMap
}

func (l *LdapDst) SelectEncryptType(t string) error {
    var err error

    switch t {
    case "md5crypt":
        l.encryptType = "{CRYPT}"
        l.hash = md5cryptFunc
    default:
        err = fmt.Errorf("not support the hash encrypt type: %s", t)
    }

    return err
}

func (l *LdapDst) Parse(pkFiled string, srcGroup *lib.EntryGroup) error {
    var err error

    searchRequest := ldaplib.NewSearchRequest(
        l.workDn,
        ldaplib.ScopeWholeSubtree, ldaplib.NeverDerefAliases, 0, 0, false,
        "(objectclass=*)",
        []string{},
        nil,
    )

    sr, err := l.conn.Search(searchRequest)
    if err != nil {
        return err
    }

    lGroup, err := lib.NewEntryGroup(pkFiled)
    if err != nil {
        return err
    }
    l.pkFiled = pkFiled

    for _, e := range sr.Entries {
        lRow, err := lib.LdapEntryToRow(pkFiled, l.syncMap, e)
        if err == lib.PrimaryFieldNotFound {
            log.Printf("ignore dn [%s] because of empty primary key\n", e.DN)
            continue
        } else if err != nil {
             return err
        }

        lRow.SetDN(e.DN)
        err = lGroup.AddRow(lRow)
        if err != nil {
            return err
        }
    }

    insert, update, delete, err := lib.EntryGroupDiff(l.syncMap, srcGroup, lGroup)
    if err != nil {
        return err
    }
    log.Printf("after LDAP entry group diff, get length of insert is: %d\n", len(insert))
    log.Printf("after LDAP entry group diff, get length of update is: %d\n", len(update))
    log.Printf("after LDAP entry group diff, get length of delete is: %d\n", len(delete))

    updateUeqList, err := l.generateOpUpdate(update)
    if err != nil {
        return err
    }
    l.opUpdate = updateUeqList

    deleteReqList, err := l.generateOpDelete(delete)
    if err != nil {
        return err
    }
    l.opDelete = deleteReqList

    insertReqList, err := l.generateOpInsert(insert)
    if err != nil {
        return err
    }
    l.opInsert = insertReqList

    return nil
}

func (l *LdapDst) Sync() error {
    var err error

    // working for insert operation
    for _, req := range l.opInsert {
        err = l.conn.Add(req)
        if err != nil {
            return fmt.Errorf("get an error when insert %s: %v", req.DN, err)
        }
    }

    // working for update operation
    for _, req := range l.opUpdate {
        err = l.conn.Modify(req)
        if err != nil {
            return fmt.Errorf("get an error when modify %s: %v", req.DN, err)
        }
    }

    // working for delete operation
    for _, req := range l.opDelete {
        err = l.conn.Del(req)
        if err != nil {
            return fmt.Errorf("get an error when delete %s: %v", req.DN, err)
        }
    }

    return err
}

func (l *LdapDst) ParsePrint() error {
    log.Println("----------> ParsePrint <----------")

    // print for insert operation
    log.Printf("########## insert operation: %d\n", len(l.opInsert))
    for _, req := range l.opInsert {
        log.Printf("DN: %s\n", req.DN)
        log.Println("OP:")
        for _, a := range req.Attributes {
            log.Printf("\t%s: %v\n", a.Type, a.Vals)
        }
    }
    log.Println("########## EOF insert operation")

    // print for update operation
    log.Printf("########## update operation: %d\n", len(l.opUpdate))
    for _, req := range l.opUpdate {
        log.Printf("DN: %s\n", req.DN)
        log.Println("OP:")
        for _, c := range req.Changes {
            log.Printf("\t%v\n", c.Modification.Vals)
        }
    }
    log.Println("########## EOF update operation")

    // print for delete operation
    log.Printf("########## delete operation: %d\n", len(l.opDelete))
    for _, req := range l.opDelete {
        log.Printf("DN: %s\n", req.DN)
    }
    log.Println("########## EOF delete operation")

    log.Println("--------> EOF ParsePrint <--------")

    return nil
}

func (l *LdapDst) generateOpInsert(rows []*lib.EntryRow) ([]*ldaplib.AddRequest, error) {
    var err error
    reqList := make([]*ldaplib.AddRequest, len(rows))

    dn := l.workDn
    pkFiled := l.pkFiled

    for idx, e := range rows {
        if dn == "" {
            return []*ldaplib.AddRequest{}, errors.New("get an empty dn from row")
        }
        dnInsert := fmt.Sprintf("%s=%s,%s",pkFiled, e.PKName(), dn)
        req := ldaplib.NewAddRequest(dnInsert, nil)
        for k, val := range e.GetRow() {
            if utils.FindStrSlice(passwdFieldSample, strings.ToLower(k)) != -1 {
                var _t []string
                for _, tt := range val {
                    hashStr, err := l.hash(tt)
                    if err != nil {
                        return []*ldaplib.AddRequest{}, err
                    }
                    _t = append(_t, l.encryptType + hashStr)
                }
                val = _t
            }
            req.Attribute(k, val)
        }
        req.Attribute("objectClass", l.usedobjectClass)
        reqList[idx] = req
    }

    return reqList, err
}

func (l *LdapDst) generateOpUpdate(rows []*lib.EntryRow) ([]*ldaplib.ModifyRequest, error) {
    var err error
    reqList := make([]*ldaplib.ModifyRequest, len(rows))

    dn := l.workDn
    pkFiled := l.pkFiled

    for idx, e := range rows {
        if dn == "" {
            return []*ldaplib.ModifyRequest{}, errors.New("get an empty dn from row")
        }
        dnUpdate := fmt.Sprintf("%s=%s,%s",pkFiled, e.PKName(), dn)
        req := ldaplib.NewModifyRequest(dnUpdate, nil)
        for k, val := range e.GetRow() {
            if utils.FindStrSlice(passwdFieldSample, strings.ToLower(k)) != -1 {
                var _t []string
                for _, tt := range val {
                    hashStr, err := l.hash(tt)
                    if err != nil {
                        return []*ldaplib.ModifyRequest{}, err
                    }
                    _t = append(_t, l.encryptType + hashStr)
                }
                val = _t
            }
            req.Replace(k, val)
        }
        reqList[idx] = req
    }

    return reqList, err
}

func (l *LdapDst) generateOpDelete(rows []*lib.EntryRow) ([]*ldaplib.DelRequest, error) {
    var err error
    reqList := make([]*ldaplib.DelRequest, len(rows))

    for idx, e := range rows {
        dn := e.GetDN()
        if dn == "" {
            return []*ldaplib.DelRequest{}, errors.New("get an empty dn from row")
        }
        req := ldaplib.NewDelRequest(dn, nil)
        reqList[idx] = req
    }

    return reqList, err
}

func (l *LdapDst) GetOpUpdate() OpUpdate {
    return l.opUpdate
}

func (l *LdapDst) GetOpInsert() OpInsert {
    return l.opInsert
}

func (l *LdapDst) GetOpDelete() OpDelete {
    return l.opDelete
}
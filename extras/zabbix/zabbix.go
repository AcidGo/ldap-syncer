package extra_zabbix

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "strings"

    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/sources/source"
    "github.com/AcidGo/ldap-syncer/utils"
)

type opUserCreate struct {
    Alias           string                  `json:"alias"`
    Usrgrps         []map[string]string     `json:"usrgrps"`
}

type opUserDelete []string

type ZabbixExtra struct {
    zapi            *ZabbixAPI
    ldapDst         *ldap.LdapDst
    ldapSA          string
    wantDel         bool
    usrgrps         []map[string]string
    userCreate      []opUserCreate
    userDelete      []opUserDelete
}

func NewZabbixExtra() (*ZabbixExtra, error) {
    return &ZabbixExtra{}, nil
}

func (ze *ZabbixExtra) BindSource(s sources.Sourcer) error {
    return nil
}

func (ze *ZabbixExtra) BindLdap(l *ldap.LdapDst) error {
    if l == nil {
        return errors.New("the binding LDAP is nil")
    }

    ze.ldapDst = l
    return nil
}

func (ze *ZabbixExtra) Parse(i interface{}) error {
    f, ok := i.(ZabbixFlags)
    if !ok {
        return errors.New("expecting extra_zabbix.ZabbixFlags")
    }

    if *f.LdapSA == "" {
        return errors.New("LDAP search attribute is emtpy for zabbix")
    }
    ze.ldapSA = *f.LdapSA

    ze.wantDel = *f.WantDel

    zapi, err := NewZabbixAPI(
        *f.URL,
        *f.User,
        *f.Passwd,
    )
    if err != nil {
        return err
    }

    ze.zapi = zapi
    _, err = ze.zapi.Login()
    if err != nil {
        return err
    }

    params := map[string]interface{} {
        "output": []string{"usrgrpid", "name"},
    }
    allUsrgrps, err := ze.zapi.UsergroupGet(params)
    if err != nil {
        return err
    }

    var usrgrps []map[string]string
    s := strings.Split(*f.Usrgrps, ",")
    for _, r := range allUsrgrps {
        if val, ok := r["name"]; ok {
            if id, ok := r["usrgrpid"]; ok && utils.FindStrSlice(s, val) != -1 {
                usrgrps = append(usrgrps, map[string]string{"usrgrpid": id})
            }
        }
    }
    if len(usrgrps) == 0 {
        return errors.New("length of final usrgrps is zero")
    }
    ze.usrgrps = usrgrps

    if ze.ldapDst == nil {
        return errors.New("the ldapDst is nil in parsing step")
    }

    ze.userCreate = make([]opUserCreate, 0)
    err = ze.generateUserCreate()
    if err != nil {
        return err
    }

    ze.userDelete = make([]opUserDelete, 0)
    if ze.wantDel {
        err = ze.generateUserDelete()
        if err != nil {
            return err
        }
    }

    return nil
}

func (ze *ZabbixExtra) ParsePrint() {
    log.Println("----------> ParsePrint <----------")

    // print for user.create API params
    log.Printf("########## user.create: %d\n", len(ze.userCreate))
    for _, i := range ze.userCreate {
        log.Printf("alias: %-20s\tusrgrps: %v\n", i.Alias, i.Usrgrps)
    }
    log.Println("########## EOF user.create")

    // print for user.delete API params
    log.Printf("########## user.delete: %d\n", len(ze.userDelete))
    for _, i := range ze.userDelete {
        log.Printf("userids: %v\n", i)
    }
    log.Println("########## EOF user.delete")

    log.Println("--------> EOF ParsePrint <--------")
}

func (ze *ZabbixExtra) Run() error {
    var err error

    // get now all user alias
    p := map[string]interface{} {
        "output": []string {"alias"},
    }
    res, err := ze.zapi.UserGet(p)
    if err != nil {
        return err
    }
    userInfolst := make([]string, 0)
    for _, i := range res {
        userInfolst = append(userInfolst, i["alias"])
    }

    // working for create user
    for _, op := range ze.userCreate {
        if utils.FindStrSlice(userInfolst, op.Alias) != -1 {
            log.Printf("zabbix user %s is not exists, ignore create\n", op.Alias)
            continue
        }
        var params map[string]interface{}
        t, _ := json.Marshal(op)
        err = json.Unmarshal(t, &params)
        if err != nil {
            return err
        }
        _, err = ze.zapi.UserCreate(params)
        if err != nil {
            return fmt.Errorf("get an error when create user %v: %v", params, err)
        }
    }

    // working for delete user
    if ze.wantDel {
        for _, op := range ze.userDelete {
            _, err = ze.zapi.UserDelete(op)
            if err != nil {
                return fmt.Errorf("get an error when delete user %v: %v", op, err)
            }
        }
    }

    return nil
}

func (ze *ZabbixExtra) generateUserCreate() error {
    opInsert, err := ze.ldapDst.GetInsert()
    if err != nil {
        return err
    }

    for _, e := range opInsert {
        m := make(map[string][]string)
        for _, a := range e.Attributes {
            m[a.Name] = a.Values
        }

        if val, ok := m[ze.ldapSA]; ok && len(val) > 0 {
            ze.userCreate = append(
                ze.userCreate, 
                opUserCreate{Alias: val[0], Usrgrps: ze.usrgrps},
            )
        } else {
            return fmt.Errorf("not %s in the map: %v", ze.ldapSA, m)
        }
    }

    return nil
}

func (ze *ZabbixExtra) generateUserDelete() error {
    opDelete, err := ze.ldapDst.GetDelete()
    if err != nil {
        return err
    }

    var ids []string
    for _, e := range opDelete {
        m := make(map[string][]string)
        for _, a := range e.Attributes {
            m[a.Name] = a.Values
        }

        if val, ok := m[ze.ldapSA]; ok && len(val) > 0 {
            params := map[string]interface{} {
                "output": []string {"userid"},
                "filter": map[string]string {"alias": val[0]},
            }
            res, err := ze.zapi.UserGet(params)
            if err != nil {
                return err
            }
            for _, i := range res {
                if userid, ok := i["userid"]; ok {
                    ids = append(ids, userid)
                }
            }
        }
    }

    if len(ids) > 0 {
        ze.userDelete = append(
            ze.userDelete,
            ids,
        )
    }

    return nil
}
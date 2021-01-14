package extra_zabbix

import (
    "errors"

    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/sources/source"
    "github.com/AcidGo/ldap-syncer/utils"
)

type opUserCreate struct {
    alias           string
    usrgrps         map[string]string
}

type opUserDelete struct {
    ids             []string
}

type ZabbixExtra struct {
    zapi            *ZabbixAPI
    ldapDst         *ldap.LdapDst
    mapping         map[string]string
    userCreate      []*opUserCreate
    userDelete      []*opUserDelete
}

func NewZabbixExtra(zURL, zUser, zPasswd string) (*ZabbixExtra, error) {
    return &ZabbixExtra{}, nil
}

func (ze *ZabbixExtra) BindSource(s source.sourcer) error {
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

    mapping := utils.StrToSyncMap(*f.Mapping)
    if len(mapping) == 0 {
        return errors.New("the zabbix user sync mapping is empty")
    }
    ze.mapping = mapping

    zapi, err := NewZabbixAPI(
        *f.URL,
        *f.User,
        *f.Passwd,
        *f.Mod,
    )
    if err != nil {
        return err
    }

    ze.zapi = zapi

    if ze.ldapDst == nil {
        return errors.New("the ldapDst is nil in parsing step")
    }

    err = ze.generateUserCreate()
    if err != nil {
        return err
    }
    err = ze.generateUserDelete()
    if err != nil {
        return err
    }

    return nil
}

func (ze *ZabbixExtra) ParsePrint() {
    return 
}

func (ze *ZabbixExtra) Run() error {

}

func (ze *ZabbixExtra) generateUserCreate() error {
    opInsert := ze.ldapDst.GetOpInsert()
    
}
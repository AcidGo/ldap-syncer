package main

import (
    "flag"
    "log"

    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/lib"
    "github.com/AcidGo/ldap-syncer/sources/mysql"
    "github.com/AcidGo/ldap-syncer/sources/source"
    "github.com/AcidGo/ldap-syncer/utils"
)

var (
    setting = src_mysql.MySQLFlags{
        ConnAddr:       flag.String("mysql-addr", "127.0.0.1:389", "MySQL listener to be connected"),
        Username:       flag.String("mysql-user", "", "MySQL connect user certificate"),
        Password:       flag.String("mysql-passwd", "", "MySQL connect user's password certificate"),
        TargetDB:       flag.String("mysql-db", "", "MySQL target database for working"),
        TargetTable:    flag.String("mysql-tb", "", "MySQL target table for working"),
    }

    ldapAddr        = flag.String("ldap-addr", "127.0.0.1:389", "LDAP listener to be connected")
    ldapBindDN      = flag.String("ldap-bind", "", "LDAP bind DN")
    ldapBindPasswd  = flag.String("ldap-passwd", "", "LDAP bind DN certificate")
    syncMapStr      = flag.String("sync-map", "", "attributes mapping when sync to LDAP")
    pkField         = flag.String("pk", "", "specified key field for selecting row")
    workingDn       = flag.String("dn", "", "into specified LDAP DN for workspace")
)

var (
    lDst        *ldap.LdapDst
    source      sources.Sourcer
    syncMap     map[string]string
    resPull     *lib.EntryGroup
    err         error
)

func main() {
    flag.Parse()
    if *ldapBindDN == "" || *ldapBindPasswd == "" || *pkField == "" || *workingDn == "" {
        log.Fatal("the args is invalid")
    }

    syncMap = utils.StrToSyncMap(*syncMapStr)
    lDst, err = ldap.NewLdapDst(*ldapAddr, *ldapBindDN, *ldapBindPasswd, *workingDn)
    if err != nil {
        log.Printf("connecting LDAP addr %s get an error", *ldapAddr)
        log.Fatal(err)
    }
    defer lDst.Close()
    log.Printf("connecting LDAP addr %s is sucessful", *ldapAddr)

    lDst.SetSyncMap(syncMap)
    log.Println("setten syncmap for dest LDAP")

    source = new(src_mysql.MySQLSrc)
    source.SetSyncMap(syncMap)
    log.Println("setten syncmap for sourcer")
    err = source.Open(setting)
    if err != nil {
        log.Println("opening source is failed")
        log.Fatal(err)
    }
    defer source.Close()

    resPull, err = source.Pull(*pkField)
    if err != nil {
        log.Println("pulling source is failed")
        log.Fatal(err)
    }
    log.Println("pulling source is sucessful")

    err = lDst.Parse(*pkField, resPull)
    if err != nil {
        log.Println("parsing dest LDAP with pulled group is failed")
        log.Fatal(err)
    }
    log.Println("parsing dest LDAP with pulled group is sucessful")

    log.Println("starting sync from source to destation ......")
    err = lDst.Sync()
    if err != nil {
        log.Println("syncing from dest to sourcer is failed")
        log.Fatal(err)
    }
    log.Println("syncing from dest to sourcer is sucessful")

    log.Println("done ......")
}
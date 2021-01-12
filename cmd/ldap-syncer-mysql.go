package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/AcidGo/ldap-syncer/ldap"
    "github.com/AcidGo/ldap-syncer/lib"
    "github.com/AcidGo/ldap-syncer/sources/mysql"
    "github.com/AcidGo/ldap-syncer/sources/source"
    "github.com/AcidGo/ldap-syncer/utils"
)

var (
    // app info
    AppName                             string
    AppAuthor                           string
    AppVersion                          string
    AppGitCommitHash                    string
    AppBuildTime                        string
    AppGoVersion                        string
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
    ldapEncryptType = flag.String("ldap-encrypt", "md5crypt", "select one encrypt algorithm for hash password")
    objectClass     = flag.String("objectclass", "", "using the objectclass for entry inserted")
    syncMapStr      = flag.String("sync-map", "", "attributes mapping when sync to LDAP")
    pkMapStr        = flag.String("pk-map", "", "specified key field for selecting row")
    workingDn       = flag.String("dn", "", "into specified LDAP DN for workspace")
    dryRun          = flag.Bool("dry-run", false, "dry-run mode, only print parsing result, not really execute")
)

var (
    lDst            *ldap.LdapDst
    source          sources.Sourcer
    usedObjectClass []string
    srcPk           string
    dstPk           string
    syncMap         map[string]string
    resPull         *lib.EntryGroup
    err             error
)

func flagUsage() {
    usageMsg := fmt.Sprintf(`%s
Version: %s
Author: %s
GitCommit: %s
BuildTIme: %s
GoVersion: %s
Options:
`, AppName, AppVersion, AppAuthor, AppGitCommitHash, AppBuildTime, AppGoVersion)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func main() {
    flag.Usage = flagUsage
    flag.Parse()
    if *ldapBindDN == "" || *ldapBindPasswd == "" || *pkMapStr == "" || *workingDn == "" {
        log.Fatal("the args is invalid")
    }
    srcPk, dstPk = utils.StrToSyncPk(*pkMapStr)
    if srcPk == "" || dstPk == "" {
        log.Fatalf("srcPk %s or dstPk %s is invalid\n", srcPk, dstPk)
    }
    usedObjectClass = strings.Split(*objectClass, ",")

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

    lDst.SetUsedObjectClass(usedObjectClass)
    log.Println("setten used objectclass for dest LDAP")

    err = lDst.SelectEncryptType(*ldapEncryptType)
    if err != nil {
        log.Println("selecting encrypt type get an error")
        log.Fatal(err)
    }
    log.Println("selected encrypt type for hasing password")

    source = new(src_mysql.MySQLSrc)
    source.SetSyncMap(syncMap)
    log.Println("setten syncmap for sourcer")
    err = source.Open(setting)
    if err != nil {
        log.Println("opening source is failed")
        log.Fatal(err)
    }
    defer source.Close()

    resPull, err = source.Pull(srcPk)
    if err != nil {
        log.Println("pulling source is failed")
        log.Fatal(err)
    }
    log.Println("pulling source is sucessful")

    err = lDst.Parse(dstPk, resPull)
    if err != nil {
        log.Println("parsing dest LDAP with pulled group is failed")
        log.Fatal(err)
    }
    log.Println("parsing dest LDAP with pulled group is sucessful")

    err = lDst.ParsePrint()
    if err != nil {
        log.Fatal(err)
    }

    if *dryRun {
        log.Println("only with dry-run mode, no execute the parsing result")
        return 
    }

    log.Println("starting sync from source to destation ......")
    err = lDst.Sync()
    if err != nil {
        log.Println("syncing from dest to sourcer is failed")
        log.Fatal(err)
    }
    log.Println("syncing from dest to sourcer is sucessful")

    log.Println("done ......")
}
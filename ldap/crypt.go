package ldap

import (
    "github.com/GehirnInc/crypt/md5_crypt"
)

type cryptFunc func (string) (string, error)

func md5cryptFunc(s string) (string, error) {
    var md5Crypt = md5_crypt.New()

    salt := []byte("$1$")
    key := []byte(s)

    out, err :=  md5Crypt.Generate(key, salt)
    if err != nil {
        return "", err
    }

    return out, nil
}
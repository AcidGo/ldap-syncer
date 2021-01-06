package utils

import (
    "strings"
)

func ParseSyncMap(smStr string) map[string]string {
    sm := make(map[string]string)
    for _, chunk := range strings.Split(smStr, ",") {
        pair := strings.Split(chunk, ":")
        if len(pair) != 2 {
            continue
        }
        k := strings.TrimSpace(pair[0])
        v := strings.TrimSpace(pair[1])
        sm[k] = v
    }

    return sm
}

func IsSameStringList(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }

    aSummary := make(map[string]int)
    bSummary := make(map[string]int)
    for _, s := range a {
        aSummary[s]++
    }
    for _, s := range b {
        bSummary[s]++
    }

    if len(aSummary) != len(bSummary) {
        return false
    }

    for aKey, aVal := range aSummary {
        if bVal, ok := bSummary[aKey]; !ok {
            return false
        } else {
            if aVal != bVal {
                return false
            }
        }
    }

    return true
}
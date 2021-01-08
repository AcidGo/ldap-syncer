package utils

import (
    "testing"
)

func TestIsSameStringList(t *testing.T) {
    data := []struct{
        input   [][]string
        output  bool
    }{
        {[][]string{[]string{"123", "456", "789"}, []string{"789", "456", "123"}}, true},
        {[][]string{[]string{"123"}, []string{"789"}}, false},
        {[][]string{[]string{"123", "123", "789"}, []string{"789", "789", "123"}}, false},
        {[][]string{[]string{"123", "123", "789", "789"}, []string{"789", "789", "123", "123"}}, true},
    }

    for i := range data {
        if res := IsSameStringList(data[i].input[0], data[i].input[1]); res != data[i].output {
            t.Errorf("IsSameStringList(%v) expected to be %v but actually was %v", data[i].input, data[i].output, res)
        }
    }
}

func TestStrToSyncMap(t *testing.T) {
    data := []struct{
        input   string
        output  map[string]string
    }{
        {
            "S1:D1,S2:D2,S3:D3", 
            map[string]string{
                "S1": "D1",
                "S2": "D2",
                "S3": "D3",
            },
        },
    }

    for i := range data {
        res := StrToSyncMap(data[i].input)
        if len(res) != len(data[i].output) {
            t.Errorf("StrToSyncMap(%v) expected to be %v but actually was %v", data[i].input, data[i].output, res)
        }
        for rKey, rVal := range res {
            if dVal, ok := data[i].output[rKey]; !ok || dVal != rVal {
                t.Errorf("StrToSyncMap(%v) expected to be %v but actually was %v", data[i].input, data[i].output, res)
            }
        }
    }
}
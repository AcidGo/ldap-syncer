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
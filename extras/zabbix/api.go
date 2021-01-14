package extra_zabbix

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "time"
)

const (
    JsonrpcVersion string = "2.0"
    JsonAuthID int = 112233
)

type ZabbixAPI struct {
    url         string
    user        string
    password    string
    id          int
    auth        string
    Client      *http.Client
}

type JsonRPCRequsetBase struct {
    Jsonrpc     string      `json:"jsonrpc"`
    Method      string      `json:"method"`
    Params      interface{} `json:"params"`
    Id          int         `json:"id"`
}

type JsonRPCRequset struct {
    Jsonrpc     string      `json:"jsonrpc"`
    Method      string      `json:"method"`
    Params      interface{} `json:"params"`
    Id          int         `json:"id"`
    Auth        string      `json:"auth"`
}

type JsonRPCResponse struct {
    Jsonrpc     string          `json:"jsonrpc"`
    Result      interface{}     `json:"result"`
    Error       ZabbixAPIError  `json:"error"`
    Id          int             `json:"id"`
}

type ZabbixAPIError struct {
    Code    int     `json:"code"`
    Message string  `json:"message"`
    Data    string  `json:"data"`
}

func NewZabbixAPI(url, user, password string) (*ZabbixAPI, error) {
    return &ZabbixAPI{
        url: url,
        user: user,
        password: password,
        auth: "",
        id: JsonAuthID,
        Client: &http.Client{
            Timeout: 150 * time.Second,
        },
    }, nil
}

func (api *ZabbixAPI) request(method string, params interface{}) (JsonRPCResponse, error) {
    id := api.id
    api.id = api.id + 1
    var err error
    var reqJson []byte
    if method != "user.login" {
        reqObj := JsonRPCRequset{
            Jsonrpc: JsonrpcVersion,
            Method: method,
            Params: params,
            Auth: api.auth,
            Id: id,
        }
        reqJson, err = json.Marshal(reqObj)
        if err != nil {
            return JsonRPCResponse{}, err
        }
    } else {
        reqObj := JsonRPCRequsetBase{
            Jsonrpc: JsonrpcVersion,
            Method: method,
            Params: params,
            Id: id,
        }
        reqJson, err = json.Marshal(reqObj)
        if err != nil {
            return JsonRPCResponse{}, err
        }
    }

    req, err := http.NewRequest("POST", api.url, bytes.NewBuffer(reqJson))
    if err != nil {
        return JsonRPCResponse{}, err
    }
    req.Header.Add("Content-Type", "application/json-rpc")

    rsp, err := api.Client.Do(req)
    if err != nil {
        return JsonRPCResponse{}, err
    }

    var res JsonRPCResponse
    var buf bytes.Buffer
    _, err = io.Copy(&buf, rsp.Body)
    if err != nil {
        return JsonRPCResponse{}, err
    }
    json.Unmarshal(buf.Bytes(), &res)

    rsp.Body.Close()

    return res, nil
}

func (api *ZabbixAPI) UserGet(params interface{}) ([]map[string]string, error) {
    rsp, err := api.request("user.create", params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    var ret []map[string]string
    res ,err := json.Marshal(rsp.Result)
    err = json.Unmarshal(res, &ret)

    return ret, err
}

func (api *ZabbixAPI) UserCreate(params interface{}) (map[string]string, error) {
    rsp, err := api.request("user.create", params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    var ret map[string]string
    res, err := json.Marshal(rsp.Result)
    err = json.Unmarshal(res, &ret)

    return ret, err
}

func (api *ZabbixAPI) UserDelete(params interface{}) (map[string]map[string][]string, error) {
    rsp, err := api.request("user.create", params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    var ret map[string]map[string][]string
    res, err := json.Marshal(rsp.Result)
    err = json.Unmarshal(res, &ret)

    return ret, err
}

func (api *ZabbixAPI) UsergroupGet(params interface{}) ([]map[string]string, error) {
    rsp, err := api.request("user.create", params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    var ret []map[string]string
    res, err := json.Marshal(rsp.Result)
    err = json.Unmarshal(res, &ret)

    return ret, err
}


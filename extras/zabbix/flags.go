package extra_zabbix

type ZabbixFlags struct {
    URL         *String
    User        *String
    Passwd      *String
    Mapping     map[string]string
}
# ldap-syncer

从不同数据源接收同步数据到 LDAP 指定 DN 中，并根据补充的外部模块对同步操作的数据做额外操作，因内部需求，目前主要实现了 MySQL 的数据源和 Zabbix 的外部模块。

### 配置

目前 `ldap-syncer-mysql` 使用命令行参数如下：

| Options                     | Description                                                  |
| --------------------------- | ------------------------------------------------------------ |
| --dry-run                   | 是否使用空运行，空运行则会打印执行操作，不会真实运行。       |
| --dn \<string\>             | 同步至 LDAP 端所选用的 DN 路径，目前只支持一次同步指定一个 DN。 |
| --objectclass \<string\>    | 对于同步时增加的 DN 时选用的 ObjectClass 属性，通过 `,` 分隔。 |
| --pk-map \<string\>         | 同步时需要分别指定源端每行的主键字段和 LDAP 的填充至 DN 的主键字段。 |
| --sync-map \<string\>       | 从源端同步到 LDAP 过程中需要转换的属性映射，如 `S1:D1,S2:D2`。 |
| --ldap-addr \<string\>      | 目标端 LDAP 的地址，格式为 \<ip\>:\<port\>。                 |
| --ldap-bind \<string\>      | LDAP 认证的绑定 DN。                                         |
| --ldap-passwd \<string\>    | LDAP 绑定 DN 的密码。                                        |
| --ldap-passwd \<string\>    | 对于 LDAP 中 DN 的密码属性所选用的加密算法，默认为 md5crypt。 |
| --mysql-addr \<string\>     | 源端 MySQL 的连接地址，格式为 \<ip\>:\<port\>。              |
| --mysql-user \<string\>     | 源端 MySQL 的登陆用户名。                                    |
| --mysql-passwd \<string\>   | 源端 MySQL 登录用户的密码。                                  |
| --mysql-db \<string\>       | 源端 MySQL 连接的数据库名。                                  |
| --mysql-tb \<string\>       | 源端 MySQL 需要同步到 LDAP 的表名。                          |
| --extra                     | 是否启用附加模块发起任务。                                   |
| --zabbix-url \<string\>     | Zabbix API 调用的地址。                                      |
| --zabbix-user \<string\>    | Zabbix API 登录的用户名。                                    |
| --zabbix-passwd \<string\>  | Zabbix API 登录使用的密码。                                  |
| --zabbix-ldapsa \<string\>  | Zabbix 使用 LDAP 认证时设定的登录关联 DN 属性。              |
| --zabbix-usrgrps \<string\> | Zabbix 对于 LDAP 同步过程中会跟随创建的用户使用的组别，如 `G1,G2,G3`。 |
| --zabbix-wantdel            | Zabbix 在 LDAP 同步过程中跟随同步用户时，对于删除动作是否会执行。 |

### 使用

这个小工具的初始需求是扩展 Zabbix 的 LDAP 认证接入范围，可以将其他 Zabbix 无法原生支持的统一认证平台，通过此工具转换为 LDAP 同步来打通与 Zabbix-Web 的交互，同时减少对 Zabbix-Web 源码的修改。

+ 编译安装

  建议 Go 1.15+，编译如下：

  ```bash
  git clone https://github.com/AcidGo/ldap-syncer.git
  go build -o bin/ldap-syncer-mysql cmd/ldap-syncer-mysql.go && bin/ldap-syncer-mysql --help
  ```

+ 使用

  可设置定期同步的计划任务来做同步，执行方式可如下：

  ```bash
  bin/ldap-syncer-mysql \
  --dn ou=Users,dc=acidgo,dc=com \
  --ldap-addr 192.168.66.131:389 \
  --ldap-passwd passwd \
  --mysql-addr 192.168.66.31:3308 \
  --mysql-db test \
  --mysql-tb test \
  --mysql-passwd suan \
  --mysql-user monitor \
  --pk-map id:cn \
  --sync-map sn:sn,uid:uid,passwd:userPassword \
  --ldap-bind cn=admin,dc=acidgo,dc=com \
  --objectclass inetOrgPerson,person,top \
  --extra \
  --zabbix-ldapsa sn \
  --zabbix-user user1 \
  --zabbix-passwd 1234567890 \
  --zabbix-url 'http://192.168.66.50/api_jsonrpc.php' \
  --zabbix-usrgrps 'Disabled' \
  ```

### 拓展

其实初始方案中通过同步 LDAP 来延伸 Zabbix 对其他统一认证平台的接入的想法还是很荒谬，可以说是实验性的，但好玩就行。


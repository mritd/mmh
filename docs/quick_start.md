# 快速开始

## 如何安装

目前官方仓库提供了 mac、linux 的预编译二进制文件，用户可以直接下载并执行安装命令即可

``` sh
# 请自行查看仓库最新 Release 版本号并替换此版本号
export MMH_VERSION=v1.5.2
# 下载
wget https://github.com/mritd/mmh/releases/download/${MMH_VERSION}/mmh_darwin_amd64
# 增加可执行权限
chmod +x mmh_darwin_amd64
# 安装(安装需要输入登录密码)
./mmh_darwin_amd64 install
```

**mmh 默认会安装到 `/usr/local/bin` 目录，如果想要安装到其他目录请使用 `--dir` 选项自行指定安装目录。**

## 如何使用

mmh 安装完成后可直接在命令行运行 `mmh` 命令预览；**此命令执行后将会自动生成样例配置文件(`~/.mmh`)。**

``` sh
➜  ~ mmh
Use the arrow keys to navigate: ↓ ↑ → ←
Select Login Server:
»  prod11: root@10.10.4.11
  prod12: root@10.10.4.12

--------- Login Server ----------
Name: prod11
User: root
Address: 10.10.4.11:22
```

在配置正确的情况下选择任意一个服务器回车即可完成登录，**不过样例配置中的服务器是无法链接的，用户需要调整配置。**

## 配置调整

**mmh 默认配置信息存储在 `~/.mmh` 目录，其中默认创建的 `default.yaml` 配置文件为服务器列表配置，调整此配置添加我们真实的服务器即可。**

``` yaml
basic:
  user: root
  password: ""
  private_key: /Users/bleem/.ssh/id_rsa
  private_key_password: ""
  port: 22
  server_alive_interval: 0s
  tmux_support: ""
  tmux_auto_rename: ""
max_proxy: 5
servers:
- name: prod11
  tags:
  - prod
...
```

由于 mmh 目前功能比较多，所以配置文件比较复杂；不过一个正常的软件配置文件对于新手来说只有一个原则: **把看不懂的都删了就行了😁。** 最后精简后的配置如下:

``` yaml
basic:
  user: root
  password: ""
  private_key: /Users/bleem/.ssh/id_rsa
servers:
- name: prod11
  user: root
  password: password
  private_key: ""
  address: 10.10.4.11
  proxy: prod12
- name: prod12
  user: root
  password: ""
  private_key: /Users/bleem/.ssh/id_rsa
  address: 10.10.4.12
  proxy: ""
```

### basic

basic 段中包含了一些默认配置，在 servers 段中填写的真实服务器如果缺少相应设置，那么默认会通过 basic 中的字段进行填充

``` yaml
basic:
  user: root
  password: "123456789"
servers:
- name: prod11
  address: 10.10.4.11
```

**以上配置中如果连接 `prod11` 服务器，那么默认用户名为 `root`，密码为 `123456789`；端口不写的情况下全局默认为 `22`。**

### servers

servers 段为一个数组结构，每一个数组元素都视为一个服务器配置；每个服务器可以进行各种自定义配置，需要注意的是: **如果同时配置了 private key 和 password，那么优先使用 private key 进行认证，且会忽略 password 认证。**

### servers.proxy

每个服务器配置中的 `proxy` 字段用于实现无限跳板功能，**proxy 字段用于指定连接本服务器前需要先跳转的机器:**

``` yaml
basic:
  user: root
  password: "123456789"
servers:
- name: prod11
  address: 10.10.4.11
  proxy: prod12
- name: prod12
  address: 10.10.4.12
  proxy: prod13
- name: prod13
  address: 10.10.4.13
```

以上配置中，连接 `prod11` 服务器时由于其指定了 `proxy` 为 `prod12`，则 mmh 将首先尝试 `prod12`；其后由于 `prod12` 指定了 `proxy` 为 `prod13`，则 mmh 接着会尝试连接 `prod13`；其实际连接顺序为:

``` sh
本地 -> prod13 -> prod12 -> prod11
```

按照此种模式配置下，mmh 会不断递归寻找需要连接的服务，从而理论上可以实现无限跳板；**不过为了安全保证(防止用户误配置)，mmh 默认只允许最多 5 个跳板机穿透(可配置)，超过 5 个将会导致连接失败。**

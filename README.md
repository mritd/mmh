# mmh

> 一个使用 Go 编写的简单多服务器登录管理小工具.

### 安装

可直接从 [release](https://github.com/mritd/mmh/releases) 页下载预编译的二进制文件，然后执行 `mmh install` (需要root 权限，自动请求 `sudo`)即可；卸载直接执行 `mmh uninstall`，卸载命令不会删除 `~/.mmh` 配置目录。

**默认安装到 `/usr/bin` 目录下，如果受权限限制无法安装，请使用 `--dir` 选项指定其他安装目录**

### 基本命令

```sh
➜  ~ mmh --help

A simple Multi-server ssh tool.

Usage:
  mmh [flags]
  mmh [command]

Available Commands:
  cp          Copies files between hosts on a network
  ctx         Change current context
  exec        Batch exec command
  go          Login single server
  help        Help about any command
  install     Install mmh
  server      Server command
  uninstall   Uninstall mmh
  version     Print version

Flags:
  -h, --help   help for mmh

Use "mmh [command] --help" for more information about a command.
```

### 配置文件

从 `v1.3.0` 版本开始，支持多配置文件切换功能；安装完成后将会自动在 `$HOME/.mmh` 下创建样例配置，默认配置文件结构如下

``` sh
➜  ~ tree .mmh
.mmh
├── default.yaml
└── main.yaml
```

#### main.yaml

主配置文件结构如下

``` yaml
context:
  default:
    config_path: ./default.yaml
context_auto_downgrade: default
context_timeout: 30m
context_use: default
context_use_time: 2018-12-02T19:50:44.681871+08:00
```

主配置文件中可以配置多个 `context`，由 `context_use` 字段指明当前使用哪个 `context`，**在每个 `context` 下执行完命令都会刷新 `context_use_time` 时间戳**；如果同时配置了 `context_timeout` 和 `context_auto_downgrade` 字段，**每次执行命令前都会检查上次使用时间距今是否超过了 `context_timeout`，如果超过则将会自动回退到 `context_auto_downgrade` 指定的 `context`**；这样可以避免长时间停留在某个重要的 `context` 上从而造成误操作(比如线上环境)

#### default.yaml

这个是真正的 SSH 配置，一般情况下其与对应的 `context` 名称相同；该配置文件样例如下

``` yaml
basic:
  user: root
  password: ""
  privatekey: /Users/mritd/.ssh/id_rsa
  privatekey_password: ""
  port: 22
  proxy: ""
maxproxy: 5
servers:
- name: prod11
  tags:
  - prod
  user: root
  password: password
  privatekey: ""
  privatekey_password: ""
  server_alive_interval: 20s
  address: 10.10.4.11
  port: 22
  proxy: prod12
- name: prod12
  tags:
  - prod
  user: root
  password: ""
  privatekey: /Users/mritd/.ssh/id_rsa
  privatekey_password: password
  server_alive_interval: 10s
  address: 10.10.4.12
  port: 22
  proxy: ""
tags:
- prod
- test
```

`basic` 段为默认配置，用于在 `servers` 段中某项配置不存在时进行填充；`servers` 段中可以配置 N 多个服务器(`server`)；每个 `server` 除了常规的 SSH 相关配置外还增加了 `proxy` 字段用于支持无限跳板(具体见下文)；`tag` 字段必须存在于在下面的 `tags` 段中，该配置主要是为了给服务器打 `tag` 方便批量复制与执行；`maxproxy` 是一个数字，用于处理当出现配置错误导致 "真·无限跳板" 情况时自动断开链接


### 自动登录 

可以使用 `mgo SERVER_NAME` 直接登录，如需交互式登录可执行 `mmh` 即可；其他相关命令如 `mms ls/add/del` 
都与添加修改服务器配置相关，请自行尝试:

![mmh](img/mmh.gif)

### 无限跳板

在每个服务器配置中可以设置一个 `proxy` 字段，当登录带有 `proxy` 字段的服务器时，工具会首先链接代理节点进行跳转；
这种能力方便于在使用跳板机的情况下无感的直连跳板机之后的主机；并且其支持无限的跳板登录，如 A、B、C 三台机器，
如果 C 的 `proxy` 设置为 B，同时 B 的 `proxy` 设置为 A，那么实际在登录 C 时，工具实际连接顺序为: `local->A->B->C`

**不要去尝试循环登录，比如 `A->B->C->A` 这种配置，工具内部已经做了检测防止产生这种 "真·无限跳板" 的情况，
默认最大跳板机数量被限制为 5 台，可通过在配置文件中增加 `maxProxy` 字段进行调整**

### 管道式批量执行

在某些情况下可能需要对某些机器执行一些小命令，工具提供了 `mec` 命令用于批量执行命令:

```sh
➜  ~ mec --help

Batch exec command.

Usage:
  exec SERVER_TAG CMD [flags]

Aliases:
  exec, mec

Flags:
  -h, --help     help for exec
  -s, --single   single server
```

在配置文件中每个服务器可以配置多个 tag，**`mec` 默认对给定的 tag 下所有机器执行命令**，如需对单个机器执行请使用 `-s` 选项；
**该命令目前支持管道处理和持续执行，比如批量执行 `tail -f` 命令等；除此之外还可以配合 grep 等命令进行自由发挥**:

**默认在批量执行模式下，每行输出前会加入当前服务器的名称前缀，在 `v1.1.0` 版本调整了颜色渲染代码，从而支持了前缀顺次颜色变换
(不会出现两个相邻服务器名字颜色一样的情况)；在单服务器下则每行输出不显示当前服务器前缀**

![mec](img/mec.gif)

### 批量复制

为了尽量方便使用，模仿了一下 `scp` 命令，增加了批量复制功能 `mcp`；批量复制支持 **本地到远端多机器的文件/目录批量复制** 和
**单一远端机器到本地的文件/目录复制**；

```sh
➜  ~ mcp --help

Copies files between hosts on a network.

Usage:
  cp [-r] FILE/DIR|SERVER_TAG:PATH SERVER_NAME:PATH|FILE/DIR [flags]

Aliases:
  cp, mcp

Flags:
  -r, --dir      useless flag
  -h, --help     help for cp
  -s, --single   single server
```

**注意: 批量复制并不可靠，请谨慎使用；并非说代码不可靠，只是相对来说对于登录等功能，即使工具出问题，也很难造成灾难性后果；
但是复制功能是有能力造成文件覆盖的，从而造成灾难性后果；所以请谨慎使用，目前只针对常规情况作了大部分测试。**

![mcp](img/mcp.gif)

### 多环境切换

考虑到同时将多个环境的配置放在同一个配置文件中会有混乱，同时也可能出现误操作的情况，`v1.3.0` 版本增加了 `context` 的概念；每个 `context` 被认为是一种环境，比如 `prod`、`test`、`uat` 等，每个环境的机器配置被分成了独立的文件以方便单独修改与加载；控制使用哪个 `context` 可以使用 `mcx use CONTEXT_NAME` 命令

``` sh
➜  ~ mcx --help

Change current context.

Usage:
  ctx [flags]
  ctx [command]

Aliases:
  ctx, mcx

Available Commands:
  help        Help about any command
  ls          List context
  use         Use context

Flags:
  -h, --help   help for ctx

Use "ctx [command] --help" for more information about a command.
```

同时又增加了 `context` 自动回退功能，即给定一个超时时间，当在给定的超时时间内无操作(不会强行断开任何命令)，下次操作将会自动回退到指定的 `context`，具体请参考上文配置段

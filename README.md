# mmh

> 一个使用 G哦编写的简单多服务器登录管理小工具.

### 主要功能

```sh
➜  ~ mmh --help

A simple Multi-server ssh tool.

Usage:
  mmh [flags]
  mmh [command]

Available Commands:
  add         Add ssh server
  cp          Copies files between hosts on a network
  del         Delete ssh server
  exec        Batch exec command
  go          Login single server
  help        Help about any command
  install     Install mmh
  ls          List ssh server
  uninstall   Uninstall mmh

Flags:
      --config string   config file (default is $HOME/.mmh.yaml)
  -h, --help            help for mmh

Use "mmh [command] --help" for more information about a command.
```

#### 服务器密码保存

默认该工具在首次运行后将会创建 `$HOME/.mmh.yaml` 样例配置；配置中可以设置服务器地址、别名、登录方式等；样例配置如下:

```yaml
servers:
- name: d24
  tags:
  - doh
  user: root
  publickey: "/Users/mritd/.ssh/id_rsa"
  address: 172.16.0.24
  port: 22
- name: d33
  tags:
  - doh
  - k8s
  user: root
  password: "password"
  address: 172.16.0.33
  port: 22
tags:
  - doh
  - k8s
```

修改配置文件后可以使用 `mgo SERVER_NAME` 直接登录，如需交互式登录可执行 `mmh` 即可；其他相关命令如 `mmh ls/add/del` 都与添加修改服务器配置相关，请自行尝试

#### 批量命令执行

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
  -s, --single   Single server
```

在配置文件中每个服务器可以配置多个 tag，**`mec` 默认对给定的 tag 下所有机器执行命令**，如需对单个机器执行请使用 `-s` 选项；
**由于批量执行功能代码层使用了 pipline 拷贝，所以支持管道处理,比如批量执行 `tail -f` 命令；再配合 grep 等命令式可以自由发挥**:

![mec](img/mec.gif)

#### 批量复制

批量执行在有些复杂命令时无法写太多，所以增加里批量复制功能；批量复制支持 **本地到远端多机器的文件/目录批量复制** 和 **单一远端机器文件/目录复制到本地**；

```sh
➜  ~ mcp --help

Copies files between hosts on a network.

Usage:
  cp FILE/DIR|SERVER_TAG:PATH SERVER_NAME:PATH|FILE/DIR [flags]

Aliases:
  cp, mcp

Flags:
  -h, --help     help for cp
  -s, --single   Single server
```

**批量复制的命令模仿的是 `scp` 的解析格式，只不过将原本的 `用户名@主机地址:端口` 替换成了 `服务器名称/服务器tag`**:

![mcp](img/mcp.gif)

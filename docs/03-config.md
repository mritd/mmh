# 配置文件

- [基础概念](#基础概念)
- [配置目录](#配置目录)
- [basic 配置段](#basic 配置段)
- [servers 配置段](#servers 配置段)
- [tags 配置段](#tags 配置段)

## 基础概念

### context

mmh 将每一个配置文件视为一个独立的 `context`，用户可以通过 `mcx` 在不同的 `context` 间切换。
**每一个 `context` 被视为一种环境的服务器集合，比如生产环境、测试环境，简而言之用户应当通过 `context`
来区分不同环境。**

### basic context

在所有 yaml 配置文件中 `basic.yaml` 配置文件是一个例外配置，**`basic.yaml` 作为其他配置文件的补充配置，
在设计时 `basic.yaml` 应当放置一些常用的与环境无关的服务器，比如无论生产还是测试环境，我都期望能快速
连接我自己私人的 vps 服务器，这时可以将私人的 vps 服务器配置加入到 `basic.yaml` 中使其在任意 `context`
都可见。**

### server

每一个 server 配置都将被视为一个单独的服务器配置，server 的 `name` 将会在 `mcs` 命令中展示，用户可以通过
server 的 `name` 来连接该服务器；每个 server 可以通过 `proxy` 配置指定连接它时需要预先连接的跳板机。

### tags

tags 描述了当前 `context` 中都有那些标记，用户可以对一个服务器打上多个标记；标记是一个组概念，通过标记
mmh 可以实现对一组服务器的批量命令执行、文件上传等操作；**每个 server 上配置的 tag 必须存在于其配置文件
的 tags 列表中，**如果一个服务器上的 tag 不在 tags 列表中声明，则认为是非法 tag。

## 配置目录

默认情况下 mmh 会在 `~/.mmh` 目录下创建两个配置文件:

- `basic.yaml`: 基础服务器配置
- `default.yaml`: 默认环境服务器配置

mmh 在启动时会扫描 `~/.mmh` 目录下所有以 `yaml` 结尾的配置文件并将其加入到 `mcx` 所显示的列表中供用户切换；
**任何时刻 mmh 服务器列表仅会显示某一个配置文件中的全部服务器以及 `basic.yaml` 中的补充服务器。**如果期望
通过其他同步工具来同步 mmh 配置，**用户可以通过为 `~/.mmh` 创建软链接的方式改变配置文件实际存储位置，或者在
shell 配置中增加 `export MMH_CONFIG_DIR='/path/to/config/dir'` 来更改 mmh 默认配置目录(推荐)。**

## basic 配置段

在每一个配置文件中都会有一个 basic 的配置段，**basic 配置段存在的意义是当前配置文件中 server 配置的某个字段
没有填写时自动使用 basic 段中的相应字段进行填充。**

``` yaml
basic:
  user: root
  password: "123456789"
servers:
- name: prod11
  address: 10.10.4.11
```

**以上配置中如果连接 `prod11` 服务器，那么默认用户名为 `root`，密码为 `123456789`；端口不写的情况下全局默认为
`22`；basic 段中其他参数含义具体请参见 servers 配置。**

## servers 配置段

每个配置文件中的 servers 段是实际可连接的服务器列表配置，servers 段是一个数组结构，每个数组元素都是一个完整的
服务配置，如果服务器配置中某些字段缺失，则 mmh 将会尝试使用 basic 段中对应配置进行填充；单个 server 的配置解
释如下:

``` yaml
servers:
# 服务器名称
- name: prod11
  # 服务器地址
  address: 10.10.4.11
  # 服务器端口
  port: 22
  # 服务器 tag 列表
  tags:
  - prod
  # 登录用户
  user: root
  # 登录密码
  password: password
  # 登录私钥
  private_key: ""
  # 如果私钥需要密码则在此填写私钥密码
  private_key_password: ""
  # 连接此服务器需要预先连接的服务器(跳板机)
  proxy: prod12
  # 每隔 n 时间发送心跳包保证 ssh 不会自动断开
  server_alive_interval: 20s

  ########### 以下部分为高级配置，具体请阅读后面的高级应用章节 ###########

  # hook_cmd 用于在登录后 hook 远端输入，以实现自动化
  hook_cmd: /usr/local/bin/hook_jump_server_select.sh
  # 配合 hook_cmd 可以读取远端输出
  hook_stdout: true
  # ssh 认证的键盘挑战 hook 脚本，用户可自行扩展实现键盘挑战登录
  keyboard_auth_cmd: /usr/local/bin/keyboard_auth_otp.sh
  # ssh 连接成功后注入 session 环境变量(需要 server sshd 调整配置)
  environment:
    ENABLE_VIM_CONFIG: "true"
    Other_KEY: "Other_String_Value"
  # 开启本地 API 支持，开启后可通过远端调用本地 api
  enable_api: true

```

## tags 配置段

在针对每个服务器配置 tag 之前，需要先在配置文件的 tags 列表中定义该 tag，定义完成后可通过批量命令执行
等操作对含有某一 tag 的所有服务器进行操作。

``` yaml
servers:
- name: prod11
  address: 10.10.1.11
  # 每个服务器可以设置多个 tag
  tags: ["prod","k8s"]
- name: prod12
  address: 10.10.1.12
  tags: ["prod","docker"]

# 预先在此定义当前配置中都包含哪些 tag
tags:
- prod
- k8s
- docker
```

以上配置中执行 `mec prod ls` 命令时，mmh 通过搜索配置文件得知 `prod11`、`prod12` 中存在 `prod` 这个 tag，
然后 mmh 将会并发的在这两台服务器上执行 `ls` 命令。

[首页](.) | [上一页](02-quick_start) | [下一页](04-usage)

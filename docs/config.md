# 配置文件

- [基础概念](#基础概念)
- [目录结构](#目录结构)
- [basic 配置](#basic 配置)
- [servers 配置](#servers 配置)
- [tags 配置](#tags 配置)

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
mmh 可以实现对一组服务器的批量命令执行、文件上传等操作；每个 server 上配置的 tag 必须存在于当前配置的
tags 列表中，如果一个服务器上的 tag 不在 tags 列表中声明，则认为时非法 tag。

## 目录结构

默认情况下 mmh 会在 `~/.mmh` 目录下创建两个配置文件:

- `basic.yaml`: 基础服务器配置
- `default.yaml`: 默认环境服务器配置

mmh 在启动时会扫描 `~/.mmh` 目录下所有以 `yaml` 结尾的配置文件并将其加入到 `mcx` 所显示的列表中供用户切换；
**任何时刻 mmh 服务器列表仅会显示某一个配置文件中的全部服务器以及 `basic.yaml` 中的补充服务器。**如果期望
通其他同步工具来同步 mmh 配置，**用户可以通过为 `~/.mmh` 创建软链接的方式改变配置文件实际存储位置，或者在
shell 配置中增加 `export MMH_CONFIG_DIR='/path/to/config/dir'` 来更改 mmh 默认扫描目录(推荐)。**

## basic 配置

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

## servers 配置

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

  # tmux 窗口标题配置(试验性)

  # 是否开启 tmux 支持，开启后 mmh 连接服务器后将会自动设置当前窗口标题为服务器名称
  tmux_support: true
  # 是否开启 tmux 自动更新窗口标题，开启后 mmh 连接服务器退出后 tmux 窗口标题将会自动 rename
  tmux_auto_rename: false
  
  # 自动切换 root 配置(试验性)

  # 是否自动切换到 root
  su_root: true
  # 切换到 root 时使用 'sudo - root' 来切换
  use_sudo: true
  # 使用 'sudo - root' 切换时是否需要输入密码
  no_password_sudo: false
  # 如果 'sudo - root' 需要密码则在此填写 root 密码
  root_password: root
```

## tags 配置

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

[首页](.) | [上一页](quick_start) | [下一页](usage)

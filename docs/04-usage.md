# 功能详解

- [命令列表](#命令列表)
- [Shell 别名](#shell-别名)
- [无限跳板](#无限跳板)
- [代理连接](#代理连接)
- [远程执行](#远程执行)
- [批量复制](#批量复制)
- [ping 检测](#ping 检测)
- [端口映射](#端口映射)

## 命令列表

mmh 目前只支持 mac 和 linux 平台，对于 windows 用户在 WSL2 ubuntu 系统下已经测试通过。**这些快捷
命令既可以通过软链接的方式提供，也可以通过 `mmh alias` 以 shell 别名的方式注入到终端配置文件中。**
目前 mmh 已包含的命令列表如下:

- **`mcx`: 用于切换 context(配置文件)**
- **`mcs`: 列出当前 context 中的服务器**
- **`mgo`: 登录某一台服务器(ssh)**
- **`mec`: 在远端服务器(或一组服务器)上执行命令**
- **`mcp`: 在本地与远端服务器(或一组服务器)之间复制文件**
- **`mping`: 针对远端服务器执行 ping 检测**
- **`mtun`: 在本地与远端服务器之间创建 tcp 隧道，支持反向创建**

如果不想创建软链接，也可以通过 `mmh alias` 命令将这些命令以 shell 别名的方式注入到终端配置文件中，
详见 [Shell 别名](#shell-别名)。

## Shell 别名

`mmh alias` 可以将 `mcp`、`mcs`、`mcx`、`mec`、`mgo`、`mping`、`mtun` 这些子命令以 shell
别名的方式注入到终端配置文件中。默认会根据 `$SHELL` 自动识别当前 shell 并写入对应的配置文件:

``` sh
# 自动识别当前 shell 并注入
mmh alias
# 同时注入 zsh 和 bash 的配置文件
mmh alias --shell zsh,bash
# 写入指定的配置文件
mmh alias --file ~/.zshrc
# 删除已经注入的别名
mmh alias --remove
```

支持的 shell 以及对应的配置文件如下:

| shell | 配置文件 |
| ----- | -------- |
| zsh | `$ZDOTDIR/.zshrc` 或 `~/.zshrc` |
| bash | macOS 优先 `~/.bash_profile`，其他平台优先 `~/.bashrc` |
| fish | `$XDG_CONFIG_HOME/fish/config.fish` 或 `~/.config/fish/config.fish` |
| sh | `~/.profile` |

注入的内容被 `# >>> mmh aliases >>>` 和 `# <<< mmh aliases <<<` 两行标记包裹，
**重复执行 `mmh alias` 只会替换标记块内的内容而不会重复追加**；如果希望自定义别名中使用的
mmh 命令(比如使用绝对路径)，可以通过 `--bin` 指定:

``` sh
mmh alias --bin /usr/local/bin/mmh
```

命令执行完成后重新打开终端或执行 `source` 加载对应配置文件即可生效。

## 无限跳板

当服务器配置中配置了 `proxy` 字段后，mmh 将通过 `proxy` 指定的服务器来连接目标服务器。用户可以通过
直接执行 `mmh` 命令在下拉列表中选择服务器或直接使用 `mgo 服务器名称` 来连接到目标服务器。

Tips:

`mmh` 命令的下拉列表支持左右方向键翻页，单页面内支持 `1~9` 数字键快速跳转；无限跳板默认最多只能穿透
5 个跳板机，可以通过 `max_proxy` 进行调整:

``` yaml
basic:
...
servers:
...
tags:
...
max_proxy: 10
```

## 代理连接

除了跳板机，mmh 也支持通过 socks5 或 http 代理连接服务器，效果等价于 OpenSSH 的
`ProxyCommand nc -X 5 -x 127.0.0.1:17000 %h %p`:

``` yaml
basic:
  # 对所有 server 生效
  socks5: 127.0.0.1:17000
servers:
- name: gpu08
  address: 10.10.4.18
- name: gpu09
  address: 10.10.4.19
  # 单个 server 也可以覆盖 basic 中的配置
  http_proxy: http://127.0.0.1:7890
```

- `socks5` 支持 `host:port` 以及需要认证的 `user:password@host:port`
- `http_proxy` 支持 `host:port` 以及 `http://user:password@host:port`
- 两者同时配置时 `socks5` 优先
- 配置完成后 `mgo`、`mec`、`mcp`、`mtun` 等所有功能都会通过代理建立连接
- 当 server 同时配置了 `proxy`(跳板机)时，跳板机使用自身的代理配置连接，目标服务器通过跳板机的
  ssh 通道直达，此时目标 server 上的 `socks5`/`http_proxy` 不再生效

连接时会在 stderr 输出实际使用的代理(代理地址中的认证信息会被隐藏)，例如:

``` sh
🧦 using socks5 proxy [127.0.0.1:17000] => gpu08(10.10.4.18:22)
```

如果当前网络不需要代理(例如已经处于目标服务器所在内网)，可以通过全局 `--no-proxy` 选项临时忽略
`socks5`/`http_proxy` 配置，直接连接目标服务器:

``` sh
mgo --no-proxy gpu08
# 等价写法
mmh --no-proxy mgo gpu08
```

`--no-proxy` 只影响 `socks5`/`http_proxy`，配置的跳板机(`proxy`)链路仍然生效。

## 远程执行

`mec` 命令可对包含某 tag 的一组服务器进行批量的命令执行:

``` sh
# 在所有包含 docker tag 的服务器上执行 docker pull nginx:latest
mec docker 'docker pull nginx:latest'
```

`mec` 默认是针对 tag 组服务器操作的，如果期望对单一服务器可以使用 `-g false` 选项:

``` sh
# 在 prod11 服务器上执行 rm -rf /*
mec prod11 'rm -rf /*' -g false
```

## 批量复制

`mcp` 命令可以实现穿过跳板机将本地文件(夹)直接复制到远端服务器，其行为模式与无限跳板类似；目前其底层
通过标准的 sftp 协议传输保证其稳定性，但是由于其复杂逻辑，目前仅仅测试了以下情况:

``` sh
# 将本地多文件复制到远端，文件列表混和包括单文件与目录
mcp dir1 dir1 file1 file2 prod11:~
# 将本地多文件批量(-g)复制到远端，文件列表包括单文件与目录
mcp dir1 dir1 file1 file2 docker:~ -g
# 将远端服务器文件/目录复制到本地
mcp prod11:~/file1 ./file1
mcp prod11:~/dir1 ./dir1
```

**需要注意的是复制行为具有文件覆盖特性，相较于其他命而言令复制命令更容易造成不可恢复的故障；所以用户
在使用本命令前应当已经充分了解其带来的风险。**

## ping 检测

ping 检测依赖于宿主机 ping 命令，当执行 `mping prod11` 时会进行两种情况的判断:

- 如果 `prod11` 服务器无需穿越跳板，则直接调用本地 ping 命令执行
- 如果 `prod11` 服务器需要穿越跳板机，则始终在穿越链路的倒数第二台机器上执行 ping 命令

## 端口映射

端口映射功能提供了简单易用的 "ssh 打洞" 功能，其整体参数分为 3 部分:

- `-l`: 左侧，即左边的地址
- `-r`: 右侧，即右边的地址
- `server`: 目标服务器

**默认情况下将右侧的端口映射到左侧:**

``` sh
# 通过 prod11 的机器将 baidu.com:80 端口映射到本地
# 本地访问 127.0.0.1:8080 实际的流量传输为:
# 本地 -> 跳板机1 -> ... 跳板机n -> prod11 -> baidu.com:80
mtun -l 127.0.0.1:8080 -r baidu.com:80 prod11
```

**使用 `--reverse` 选项可以将链路反转:**

``` sh
# 通过本地将 baidu.com:80 发布到 prod11 的 8080 端口 
# 远端 prod11 会监听 127.0.0.1:8080(netstat 查看监听进程为 ssh)
# 远端 prod11 访问 127.0.0.1:8080 会将流量穿透回本地然后发出
# prod11 -> 跳板机n -> 本地 -> baidu.com:80
mtun -l baidu.com:80 -r 127.0.0.1:8080 --reverse prod11
```

[首页](.) | [上一页](03-config) | [下一页](05-advanced)

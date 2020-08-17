# 使用案例

- [持续监测日志](#持续监测日志)
- [本地连接生产 mysql](#本地连接生产 mysql)
- [生产使用本地网络](#生产使用本地网络)
- [延迟通知](#延迟通知)
- [tmux 集成](#tmux 集成)

## 持续监测日志

由于 mmh 的 exec 实现了管道式的流处理，所以利用这个功能可以实现持续的多服务器日志监测

[![asciicast](https://asciinema.org/a/ImFLQGWlUgeEtOwJOPxxxxKMk.svg)](https://asciinema.org/a/ImFLQGWlUgeEtOwJOPxxxxKMk)

## 本地连接生产 mysql

在某些时候我们期望通过本地的 GUI 工具来连接生产 mysql 服务器并执行一些 sql，但是大部分生产环境的 mysql
都禁止公网直接连接，此时可以借助 mtun 命令创建 tcp 隧道来连接生产 mysql。

``` sh
# -l 左侧表示本地监听的端口
# -r 填写与目标服务器(prod11)相同内网的 mysql 地址
# 当连接本地 3306 时，远端 prod11 会将流量转发到同内网的 172.16.3.33:3306 上
mtun -l 127.0.0.1:3306 -r 172.16.3.33:3306 prod11
```

## 生产使用本地网络

很多时候我们需要在远端生产服务器下载一些文件，往往这些文件全部发布在 GitHub 上，由于众所周知的
不可描述的原因，导致 GitHub 国内访问极其缓慢；此时可以通过 mtun 将本地的某些不可描述的服务发布
到远端服务器上，然后远端服务器下载时通过这个端口将流量路由回本地再通过某些不可描述的服务发出。

``` sh
# -l 代表本地要连接的端口(不可描述服务监听在此端口)
# -r 代表 prod11 上需要监听的端口
# --reverse 选项让 mtun 反转
# 此时在 prod11 上可以通过 127.0.0.1:8234 来转发流量
mtun -l 127.0.0.1:8123 -r 127.0.0.1:8234 prod11 --reverse
```

## 延迟通知

当在远端服务器执行一个耗时命令时，我们通常不想一直等待，只想在执行完成得道一个通知，此时可以借助
noti api 完成(请确保本地已安装 [noti](https://github.com/variadico/noti)):

``` sh
docker pull debian:10 && curl -X POST ${MMH_API_ADDR}/noti --data-binary "pull debian:10 success!"
```

当 `docker pull debian:10` 执行完成后我们会在本地弹出一个 `pull debian:10 success!` 的通知。

## tmux 集成

默认情况下 mmh 会自动检测当前是否处于 tmux 环境，如果在 tmux 中运行 mmh 则 **mmh 将自动设置连接的
server name 为本地 tmux 的 window name 以方便区分。**在新版本的 tmux 中增加了 `client-session-changed`
hook，通过此 hook + 多 tmux session 可以让 mmh 自动切换配置文件:

``` sh
# init session
new -s local
new -s bleem
new -s test
new -s prod

set-hook -t local client-session-changed 'run-shell "mcx set basic"'
set-hook -t bleem client-session-changed 'run-shell "mcx set basic"'
set-hook -t test client-session-changed 'run-shell "mcx set test"'
set-hook -t prod client-session-changed 'run-shell "mcx set prod"'
```

在以上 tmux 配置中，tmux 打开后默认初始化 4 个 session，每次切换 session 窗口后 mmh 自动切换配置文件，
这样实现了只在特定 session 操作特定服务器，防止出现 tmux 窗口各种切换然后误操作的情况。本人自用的
tmux 配置保存在 https://github.com/mritd/config/blob/master/tmux/tmux.conf 中，有需要的可以参考。

[首页](.) | [上一页](06-build) | [下一页](08-q_a)

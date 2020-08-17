# 高级应用

- [键盘挑战认证](#键盘挑战认证)
- [登录后 hook](#登录后 hook)
- [环境变量注入](#环境变量注入)
- [API 访问](#API 访问)

## 键盘挑战认证

在大部分 ssh 登录时一般都会采用用户名密码、或者公钥认证，但是在登录一些跳板机时由于其
特殊性，可能会要求用户输入 OTP 口令等；这部份 OTP 口令一般通过键盘交互挑战的方式来完成
认证；针对此种情况从 v1.5.3 版本起，mmh 提供了 `keyboard_auth_cmd` 配置用来处理键盘挑战。

`keyboard_auth_cmd` 配置需要指定一个命令，该命令可以是一个可执行的脚本或是其他的二进制程序:

``` yaml
servers:
- name: test
  address: 172.16.4.110
  port: 521
  user: bleem
  password: "1234----"
  keyboard_auth_cmd: "/etc/mmh/keyboard.sh"
```

指定此配置后，在 mmh 接收到 KeyboardInteractive 挑战时则会调用此命令；**mmh 将 KeyboardInteractive
请求以 json 形式发送到 `keyboard_auth_cmd` 的 stdin，同时接收 `keyboard_auth_cmd` 的 stdout 作为返
回结果，针对返回结果 mmh 会按行分割为多个响应返回给 ssh server。**用户可以通过此配置 hook 任意的
键盘挑战，常见的 OTP 响应可以通过一些命令行工具完成(例如 [otp-authenticator](https://github.com/mstksg/otp-authenticator))。

## 登录后 hook

在登录成功后用户往往可能期望执行一些自定义操作，比如自动选择跳板机服务器、通过 `sudo` 切换到
root 用户等等；针对这种情况从 v1.5.3 版本起 mmh 提供了 `hook_cmd` 配置来控制标准输入输出完成
自动化操作:

``` yaml
servers:
- name: test
  address: 172.16.4.110
  port: 521
  user: bleem
  password: "1234----"
  keyboard_auth_cmd: "/etc/mmh/keyboard.sh"
  hook_cmd: "/etc/mmh/hook_sudo_su_root.sh"
  hook_stdout: false
```

同样的 `hook_cmd` 也需要提供一个命令，当 `hook_stdout` 设置为 true 时，ssh 登录完成后 ssh server 的
stdout 输出将会被发送到 `hook_cmd` 的 stdin 中，同时 `hook_cmd` 的 stdout 将会作为命令完整的写入
ssh server 的 stdin；**简单的说就是 `hook_cmd` 标准输出都会作为命令发送到 ssh server，**下面的脚本
展示了如何在有密码的情况下自动 sudo 切换 root:

``` sh
# 发送切换命令
echo "sudo su - root && exit"
# 延迟 0.5s 发送密码
sleep 0.5
echo "password"
```

## 环境变量注入

从 v1.5.4 版本开始 mmh 支持在 server 配置中加入环境变量；mmh 连接到远端服务器后会尝试将环境变量注入
到当前 ssh session 中；**需要注意的是此功能需要服务端 sshd 配置调整，默认情况下大部分发行版只支持 `LANG`、
`LC_*` 变量的注入，如果想要注入自定义环境变量请在 sshd 配置中增加对应变量:**

``` sh
# /etc/ssh/sshd_config
# Allow client to pass locale environment variables
AcceptEnv LANG LC_* ENABLE_VIM_CONFIG MMH*
```

以上配置中允许 ssh client 注入 `ENABLE_VIM_CONFIG`、`MMH*` 变量。

## API 使用

为了满足更强大的扩展能力，自 v1.5.4 版本开始 server 配置中增加了 `enable_api` 选项；当此配置开启后
**mmh 连接目标服务器成功后会随机在目标服务器上监听一个本地 http 端口，并将其请求转发到本地的 mmh
上，本地 mmh 响应对应请求并完成处理；远端服务器上可以通过 `${MMH_API_ADDR}` 变量获取到监听地址(需要
环境变量注入支持)。**在当前版本仅存在几个内置的 API 集成，后续会考虑开放给用户进行自定义和完善；
当前支持的 API 如下:

- `${MMH_API_ADDR}/`: 单纯的返回一段字符串代表 mmh api server 已经运行
- `${MMH_API_ADDR}/healthz`: mmh api server 健康检测接口
- `${MMH_API_ADDR}/copy`: 通过 POST 请求，body 内的内容会自动复制到本地剪切板(只支持纯文本)
- `${MMH_API_ADDR}/noti`: 通过 POST 请求，如果本地安装了 [noti](https://github.com/variadico/noti) 则 body 内容作为 message 在本地弹出

`copy` api 一般用于在无限远端想要复制一个大文本，拖屏幕很不方便的时候可以通过 api 完成，以下
为一个小脚本用来模仿 mac 本地的 `pbcopy` 命令:

> 假设将脚本保存为 `mcopy` 文件，在远端执行 `cat ~/.zshrc | mcopy` 将会直接复制到本地剪切板。

``` sh
#!/usr/bin/env bash

if [ ! -n "${MMH_API_ADDR}" ]; then
    echo "MMH API is not enabled, or sshd does not support injection of MMH_API_ADDR env."
    exit 1
fi

curl -X POST ${MMH_API_ADDR}/copy --data-binary @-
```

`noti` api 作为 noti 项目的增强，一般用于在无限远端执行一个耗时命令但不想一直等待，此时可以
通过 `noti` api 跨越跳板机进行回调通知，以下为一个模拟本地 noti 的小脚本:

> 假设将脚本保存为 `mti`，在远端执行 `sleep 30 && mti This is a test noti.` 将会在 30s 后收到通知。

``` sh
#!/usr/bin/env bash

if [ ! -n "${MMH_API_ADDR}" ]; then
    echo "MMH API is not enabled, or sshd does not support injection of MMH_API_ADDR env."
    exit 1
fi

curl -X POST ${MMH_API_ADDR}/noti --data-binary "$*"
```

[首页](.) | [上一页](04-usage) | [下一页](06-build)

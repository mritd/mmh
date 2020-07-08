# 高级应用

- [键盘挑战认证](#键盘挑战认证)
- [登录后 hook](#登录后 hook)

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

[首页](.) | [上一页](usage) | [下一页](build)

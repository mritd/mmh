# 自行编译

当前源码编译依赖环境如下:

- Go 1.14.2+
- Make
- [gox](https://github.com/mitchellh/gox)

在环境配置完成后在本项目根目录下执行编译即可

``` sh
# 交叉编译，编译完成后文件输出在 dist 目录
make

# 仅仅安装到 ${GOPATH}/bin 中
make install
```

[首页](.) | [上一页](usage) | [下一页](example)

# 自行编译

当前源码编译依赖环境如下:

- Go 1.14.2+
- Make
- [gox](https://github.com/mitchellh/gox)

在环境配置完成后在本项目根目录下执行编译即可

``` sh
# 编译(编译文件在 dist 目录，默认交叉编译所有平台)
make

# 仅仅安装到 ${GOPATH}/bin
make install
```

### Promptx

> 一个简单的 golang 编写的终端提示库，部分代码 copy 自 [promptui](https://github.com/manifoldco/promptui)，代码设计实现参考了 [promptui](https://github.com/manifoldco/promptui)
与 [survey](https://github.com/AlecAivazis/survey)；至于为何不使用这两个库而自己造轮子是因为这两个库对中文支持都不友好，比如光标移动、删除等

**Promptx 使用可参考 [_example](https://github.com/mritd/promptx/tree/develop/_example) 目录下的样例代码，目前仅支持单行提示(可以增加实时校验)
与下来列表选择菜单，其运行截图如下:**

#### 单行提示

![prompt](img/prompt.gif)

#### 列表选择

![select](img/select.gif)
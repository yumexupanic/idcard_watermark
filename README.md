## idcard_watermark

由于一些第三方平台需要身份证信息来做实名认证 虽然我们无法保证第三方平台是不是会出卖用户数据 但是为了自己的身份证信息不会被滥用 可以考虑给身份证添加水印。

效果图如下:

![image](https://s2.ax1x.com/2019/05/04/EaQ97n.png)

## 快速开始(Quick Start)

```shell
go build src/main.go
./main -target ~/a.png
./main -target ~/a.png -output aaa.png -text 123123 -fonts fonts/SourceHanSansK-Normal.ttf
./main -target ./imgs/
```
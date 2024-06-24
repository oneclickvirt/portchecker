# portchecker

[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fportchecker&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com) [![Build and Release](https://github.com/oneclickvirt/portchecker/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/portchecker/actions/workflows/main.yaml)

端口检测模块 (port checker)

## 功能(Features)

- [x] 本机邮件常用端口检测
- [x] 常用邮件平台的SMTP、POP3、IMAP协议检测

## TODO

- [ ] 常用端口多地区是否阻断检测 

## 使用(Usage)

下载及安装

```
curl https://raw.githubusercontent.com/oneclickvirt/portchecker/main/pck_install.sh -sSf | sh
```

使用

```
pck
```

或

```
./pck
```

进行测试

```

```

![图片](https://github.com/oneclickvirt/portchecker/assets/103393591/666f231b-09ba-4c3f-8cf9-c30f43365ddf)

## 在Golang中使用

```
go get github.com/oneclickvirt/portchecker@latest
```
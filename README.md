# portchecker

[![Hits](https://hits.spiritlhl.net/portchecker.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net)

[![Build and Release](https://github.com/oneclickvirt/portchecker/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/portchecker/actions/workflows/main.yaml)

端口检测模块 (port checker)

## 功能(Features)

- [x] 本机邮件常用端口检测
- [x] 常用邮件平台的SMTP、POP3、IMAP协议检测
- [x] 部分Windows10系统下打勾打叉编码错误显示，已判断是Win时使用Y/N显示而不是勾叉

## TODO

- [ ] 常用端口多地区是否阻断检测
- [ ] 支持TCP/UDP协议分开检测是否开通

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

![image](https://github.com/user-attachments/assets/a6aeebf1-c5b1-4a18-91ac-8242f2107ec5)

## 卸载

```
rm -rf /root/pck
rm -rf /usr/bin/pck
```

## 在Golang中使用

```
go get github.com/oneclickvirt/portchecker@v0.0.3-20250629044850
```

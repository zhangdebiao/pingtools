# pingtools

指定某个ip(段)，持续ping，一旦有通企微机器人通知。
```
pingtools.exe -h 192.168.1.123
```

支持以服务方式安装运行(默认服务名Network)，需要管理员权限。
```
pingtools.exe -h 192.168.1.123/24 -i
net start Network
services Network start
```

## 参考自https://www.nodeseek.com/post-183613-1 ，在此感谢大佬的教程

说明：在大佬的基础上添加了检测更新和重启服务等功能

## 脚本界面预览：

```
欢迎使用realm一键转发脚本
=================
1. 部署环境
2. 添加转发
3. 删除转发
4. 启动服务
5. 停止服务
6. 一键卸载
7. 检测更新
8. 重启服务
0. 退出脚本
=================
realm 状态：已安装
realm 转发状态：启用
请选择一个选项: 
```
## 一键脚本：
国内或v6 only可用（推荐）
```
curl -L https://host.wxgwxha.eu.org/https://raw.githubusercontent.com/wcwq98/realm/refs/heads/main/realm.sh -o realm.sh && chmod +x realm.sh && sudo ./realm.sh
```
或
```
curl -L https://raw.githubusercontent.com/wcwq98/realm/refs/heads/main/realm.sh -o realm.sh && chmod +x realm.sh && sudo ./realm.sh
```




添加转发的配置文件：
```
nano /root/realm/config.toml
```
添加你的转发内容：
```
[[endpoints]]
listen = "0.0.0.0:本地监听端口"
#如果是v6，把 0.0.0.0 改为 [::]
remote = "需要转发的ip或域名:目标端口"
[[endpoints]]
listen = "0.0.0.0:示例2端口"
remote = "示例2ip或域名:示例2目标端口"
```
最后Ctrl O保存，Ctrl X退出

## 如需其他更多配置请参考官方示例配置： https://github.com/zhboner/realm/tree/master/examples

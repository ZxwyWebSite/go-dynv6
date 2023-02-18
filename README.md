# go-dynv6
dynv6 DDNS 的 golang 版更新工具  
依赖服务：https://dynv6.com/
## 前言
为了给海思盒子做ipv6 DDNS，所以写了这个程序
## 警告
第一次写Golang程序，可能有亿些问题
## 更新
\> 20230218  
重构代码，增加提示。  
可修改计划任务间隔时间。  
...
## 用法
第一次执行会在当前目录生成conf.ini配置文件  
token填你的dynv6 Api Token  
domain填你的DDNS域名  
ip一般不用动，填disable禁用  
wait是间隔时间，单位 秒  
show是显示当前ip和配置ip的对比，填false禁用
## 其它
随缘更新。

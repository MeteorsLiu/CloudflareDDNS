# CloudflareDDNS
A Cross-platform DDNS software for Cloudflare




# 中文文档

二进制文件自行翻release，已经上传
目前兼容平台：

Apple M1(Darwin ARM64)

MacOS(Darwin)

Linux ARM64

Linux AMD64

Linux MIPS64

Windows AMD64


使用方法：

./ddns -key xxxxx -email xxxx -domain xx.xx.com -hook xxxx -time 1


参数解释：

## time
多长时间检测一次ip是否变化（默认15秒

## key
Cloudflare API Key

## email
你的Cloudflare邮箱

## hook
每次ip变化后执行的bash shell（仅支持Linux Bash

## verbose
是否开启程序输出显示(默认：true: 开启，false: 关闭)

## domain
指定的DDNS域名

例如
`xxsad.123123.ghl.info`

**务必要先在cloudflare里面添加该域名及前缀，否则无法运行，本程序不会主动帮你添加！！**

## dev
指定interface口，仅在lan模式下生效


## timeout
查询当前IP超时时间(默认: 15秒)

## query
指定查询当前IP的URL, 忽略下面的mode

## mode
**akamai**: 使用(whatismyip.akamai.com)来查询当前外网IP

**china**: 使用myip.ipip.net(默认：akamai)

**lan**: 不查询外网IP，仅同步指定interface口(见dev参数)内网IP


mode对于那些不方便访问国外的国内机器来说很有用

而lan则是对于办公室需要使用内网动态ip环境搭建http服务器很有用





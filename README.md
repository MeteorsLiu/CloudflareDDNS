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

**time: 多长时间检测一次ip是否变化（默认1秒）**

**key：Cloudflare API Key**

**email：你的Cloudflare邮箱**

**hook: 每次ip变化后执行的bash shell（仅支持Linux Bash）**

**domain: 指定的DDNS域名，** 例如

xxsad.123123.ghl.info

**务必要先在cloudflare里面添加该域名及前缀，否则无法运行，本程序不会主动帮你添加！！**




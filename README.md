# Email-SSL
代码在阿里云服务器和阿里云邮箱服务下，已通过测试.
使用阿里云的服务发送邮件必须是 https，SSL 传输. http 的 25 端口已被禁用。
参考链接：https://help.aliyun.com/document_detail/29449.html
端口，465;
host, smtp.mxhichina.com
发件人和 email.Username 应该是一致的，否则阿里云服务器会抛出异常.   
# 例子
```$xslt    
usage:
	email = Email{
	Host:     Conf["host"],
	Port:     Conf["port"],
	Username: Conf["username"],
	Password: Conf["password"],
	}
 email.SSLSend("收件人", "主题", "正文")
```


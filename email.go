package email

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

//代码在阿里云服务器和阿里云邮箱服务下，已通过测试.
//使用阿里云的服务发送邮件必须是 https，SSL 传输. http 的 25 端口已被禁用。
//参考链接：https://help.aliyun.com/document_detail/29449.html
//端口，465;
//host, smtp.mxhichina.com
//
//usage:
//	email = Email{
//	Host:     Conf["host"],
//	Port:     Conf["port"],
//	Username: Conf["username"],
//	Password: Conf["password"],
//	}
// email.SSLSend("收件人", "主题", "正文")
//
//发件人和 email.Username 应该是一致的，否则阿里云服务器会抛出异常.

type Email struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (e *Email) SSLSend(toEmail string, subject string, content string) error {
	from := mail.Address{Address: e.Username}
	to := mail.Address{Address: toEmail}
	subj := subject
	body := content

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%s", e.Host, e.Port)

	host, _, _ := net.SplitHostPort(servername)
	log.Println("e.Username=>", e.Username, from)
	auth := smtp.PlainAuth("", e.Username, e.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func (e *Email) SendAttachSSL(toEmail string, fromEmail string, subject string, content string, attachName string, attachPath string) error {
	var (
		serverAddr = e.Host
		password   = e.Password
		emailAddr  = e.Username
		portNumber = 465
		tos        = []string{
			toEmail,
		}
		cc                 []string
		attachmentFilePath = attachPath
		filename           = attachName
		delimeter          = "**=myohmy689407924327"
	)

	tlsConfig := tls.Config{
		ServerName:         serverAddr,
		InsecureSkipVerify: true,
	}

	log.Println("Establish TLS connection")
	conn, connErr := tls.Dial("tcp", fmt.Sprintf("%s:%d", serverAddr, portNumber), &tlsConfig)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()

	log.Println("create new email client")
	client, clientErr := smtp.NewClient(conn, serverAddr)
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()

	log.Println("setup authenticate credential")
	auth := smtp.PlainAuth("", emailAddr, password, serverAddr)

	if err := client.Auth(auth); err != nil {
		return err
	}

	log.Println("Start write mail content")
	log.Println("Set 'FROM'")
	if err := client.Mail(emailAddr); err != nil {
		return err
	}
	log.Println("Set 'TO(s)'")
	for _, to := range tos {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}

	writer, writerErr := client.Data()
	if writerErr != nil {
		return writerErr
	}

	//basic email headers
	sampleMsg := fmt.Sprintf("From: %s\r\n", emailAddr)
	sampleMsg += fmt.Sprintf("To: %s\r\n", strings.Join(tos, ";"))
	if len(cc) > 0 {
		sampleMsg += fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ";"))
	}
	sampleMsg += "Subject: " + subject + "\r\n"

	log.Println("Mark content to accept multiple contents")
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)

	//place HTML message
	log.Println("Put HTML message")
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"
	sampleMsg += fmt.Sprintf("\r\n%s", "<html><body><h1>"+content+"</h1>")

	//place file
	log.Println("Put file attachment")
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: base64\r\n"
	sampleMsg += "Content-Disposition: attachment;filename=\"" + filename + "\"\r\n"
	//read file
	rawFile, fileErr := ioutil.ReadFile(attachmentFilePath)
	if fileErr != nil {
		return errors.New(fmt.Sprintf("attachmentFilePath %s is invalid", attachmentFilePath))
	}
	sampleMsg += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)

	//write into email client stream writter
	log.Println("Write content into client writter I/O")
	if _, err := writer.Write([]byte(sampleMsg)); err != nil {
		return err
	}

	if closeErr := writer.Close(); closeErr != nil {
		return closeErr
	}

	client.Quit()

	log.Print("done.")
	return nil
}

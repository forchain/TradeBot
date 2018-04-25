package helpers

import (
	"gopkg.in/gomail.v2"
	log "github.com/sirupsen/logrus"
	"fmt"
)
type MailInfo struct {
	Adrress string
	Password string
	SMTP string
	SMTPPort int
}
var MInfo *MailInfo

func DispatchAbnormalExitNotice(attachment string){
	log.Infof("发送邮件：%+v",*MInfo)
	msg:=fmt.Sprintf("😤Fuck! 噢 我非法退出了 ！😤😤😤😤😤 <br>你可以查看日志文件 在附件里👹")
	SendMail(msg,attachment)
}
func SendMail(msg string,attachment string) error{
	if MInfo.Adrress=="" {
		log.Info("没有配置邮箱")
		return nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", MInfo.Adrress)
	m.SetHeader("To", MInfo.Adrress)
	m.SetHeader("Subject", "我是 自动化交易程序 👻✌️")
	m.SetBody("text/html", msg)
	if attachment!="" {
		m.Attach(attachment)
	}


	d := gomail.NewDialer(MInfo.SMTP, MInfo.SMTPPort, MInfo.Adrress, MInfo.Password)
	err := d.DialAndSend(m)
	if  err != nil {
		log.Errorf("发送邮件错误：%v",err)
	}
	return err
}
package emailwrapper

import (
	"crypto/tls"
	"net/smtp"

	"github.com/jordan-wright/email"
)

var from = ""
var to = []string{""}
var password = ""

func Send(attachment string) error {
	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = "Hubble Report"
	e.Text = []byte("This email is automatically sent by Hubble.")
	e.AttachFile(attachment)

	var emptyConfig tls.Config
	emptyConfig.ServerName = "smtp.qiye.aliyun.com"

	err := e.SendWithTLS("smtp.qiye.aliyun.com:465", smtp.PlainAuth("", "zhaoyushi@wxblockchain.com", password, "smtp.qiye.aliyun.com"), &emptyConfig)
	return err
}

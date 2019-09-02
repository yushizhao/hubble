package emailwrapper

import (
	"crypto/tls"
	"net/smtp"

	"github.com/jordan-wright/email"
)

var from = ""
var to = []string{""}
var password = ""

func Send(attachments []string) error {
	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = "Hubble Report (TEST)"
	e.Text = []byte("This email is automatically sent by Hubble.")

	for _, attachment := range attachments {
		_, err := e.AttachFile(attachment)
		if err != nil {
			return err
		}
	}

	var emptyConfig tls.Config
	emptyConfig.ServerName = "smtp.qiye.aliyun.com"

	err := e.SendWithTLS("smtp.qiye.aliyun.com:465", smtp.PlainAuth("", "zhaoyushi@wxblockchain.com", password, "smtp.qiye.aliyun.com"), &emptyConfig)
	return err
}

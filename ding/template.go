package ding

var ErrorMsgTemplate = `{"title":"ERROR","text":"# %v"}`
var InvitationCodeTemplate = `{"title":"Invitation Code","text":"# New Invitation Code\n%d"}`

var MarkdownJsonTemplate = map[string]interface{}{
	"msgtype": "markdown",
	"at":      map[string]interface{}{"isAtAll": false},
}

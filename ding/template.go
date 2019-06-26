package ding

var ErrorMsgTemplate = `{"title":"ERROR","text":"# %v"}`

var MarkdownJsonTemplate = map[string]interface{}{
	"msgtype": "markdown",
	"at":      map[string]interface{}{"isAtAll": false},
}

package msg

// 消息体
// Msg defines message body
type Msg struct {
	Channels []string
	Content  []byte // json content
}

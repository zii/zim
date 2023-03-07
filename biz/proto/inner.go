package proto

// 内部服务包

type CmdDisconnect struct {
	UserId string `json:"user_id"`
	Token  string `json:"token"`
	Reason string `json:"reason"`
}

type Command struct {
	Op            string         `json:"op"`
	UserIds       []string       `json:"user_ids"`
	Message       *Message       `json:"message"`
	CmdDisconnect *CmdDisconnect `json:"cmd_disconnect,omitempty"`
}

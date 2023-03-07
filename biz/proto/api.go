package proto

import (
	"encoding/json"

	"zim.cn/biz/def"
)

// 授权结果
type Authorization struct {
	Token string `json:"token"` // 授权token
}

// 用户简单信息
type TinyUser struct {
	Id   string `json:"id"`   // 用户ID
	Name string `json:"name"` // 用户昵称
}

// 用户基本信息
type User struct {
	Id     string `json:"id"`     // 用户ID
	Name   string `json:"name"`   // 用户昵称
	Photo  string `json:"photo"`  // 用户头像
	Ex     string `json:"ex"`     // 用户扩展信息
	Status int    `json:"status"` // 用户状态(0正常 1禁用)
}

// 群基本信息
type Chat struct {
	Id      string `json:"id"`       // 群ID
	Type    int    `json:"type"`     // 群类型, 1普通群, 2超级群
	Title   string `json:"title"`    // 群标题
	About   string `json:"about"`    // 群公告
	OwnerId string `json:"owner_id"` // 群主ID
	Photo   string `json:"photo"`    // 群图标
	Maxp    int    `json:"maxp"`     // 人数上限, 0无限
	Muted   bool   `json:"muted"`    // 是否全员禁言
}

// 阿里oss直传凭证
// @Description
type OssCred struct {
	Endpoint        string `json:"endpoint" example:"oss-cn-guangzhou.aliyuncs.com"`                         // oss节点域名
	AccessKeyId     string `json:"access_key_id" example:"STS.NTp74bN1Z5JNgjjqhqnZPqVYJ"`                    // 临时key
	AccessKeySecret string `json:"access_key_secret" example:"8acGDXgdLmPNcZmmoKEKmeCqjQz5SVWnYqq556YKL26d"` // 临时secret
	Token           string `json:"token"`                                                                    // 上传用的token, 较长
	Timeout         int    `json:"timeout" example:"3600"`                                                   // 过期秒数, 默认3600
	Bucket          string `json:"bucket" example:"im1"`                                                     // 桶名
	FinalHost       string `json:"final_host" example:"http://openim1.oss-accelerate.aliyuncs.com"`          // 最终文件URL前缀
}

// 华为obs直传凭证
// @Description 参考文档: https://support.huaweicloud.com/bestpractice-obs/obs_05_1203.html
type ObsCred struct {
	SignedUrl string `json:"signed_url"` // 预签名URL
	Timeout   int    `json:"timeout"`    // 过期秒数, 默认3600
}

// 腾讯云cos直传凭证
// @Description 参考文档: https://cloud.tencent.com/document/product/436/9067
type CosCred struct {
	Bucket       string `json:"bucket" example:"cat1-1254751699"` // 桶名
	Region       string `json:"region" example:"ap-nanjing"`      // 区域名
	TmpSecretId  string `json:"secret_id"`                        // 临时key
	TmpSecretKey string `json:"secret_key"`                       // 临时secret
	SessionToken string `json:"token"`                            // sessionToken
	Timeout      int    `json:"timeout"`                          // 过期秒数, 默认3600
}

// Minio直传凭证
type MinioCred struct {
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	FinalHost string `json:"final_host"`
	AccessId  string `json:"access_id"`
	AccessKey string `json:"access_key"`
	Token     string `json:"token"`
	Timeout   int    `json:"timeout"`
}

// 直传凭证
type Credential struct {
	Platform string     `json:"platform"` // 平台类型(oss阿里云,obs华为云,cos腾讯云,minio)
	Oss      *OssCred   `json:"oss,omitempty"`
	Obs      *ObsCred   `json:"obs,omitempty"`
	Cos      *CosCred   `json:"cos,omitempty"`
	Minio    *MinioCred `json:"minio,omitempty"`
}

// 图片消息格式
type PhotoElem struct {
	Big   string `json:"big"`   // 原图URL
	Small string `json:"small"` // 缩略图URL
}

// 音频消息格式
type SoundElem struct {
	Url string `json:"url"` // 音频文件URL
}

// 视频消息格式
type VideoElem struct {
	Url string `json:"url"` // 视频文件URL
}

// 文件消息格式
type FileElem struct {
	Url string `json:"url"` // 文件URL
}

// @消息格式
type MentionElem struct {
	Text  string      `json:"text"`  // 常规消息文本
	Users []*TinyUser `json:"users"` // 被提及的用户列表
}

// 位置消息格式
type LocationElem struct {
	Long float64 `json:"long"` // 经度
	Lat  float64 `json:"lat"`  // 纬度
}

// 引用消息
type QuoteElem struct {
	Text string   `json:"text"` // 常规消息文本
	Msg  *Message `json:"msg"`  // 引用的消息
}

// 撤消消息(合并转发用到)
type RevokeElem struct {
	MsgId int64 `json:"msg_id"` // 消息ID
}

// 聊天记录
type ChatLogElem struct {
	Title string     `json:"string"` // XX的聊天记录
	Msgs  []*Message `json:"msgs"`   // 消息列表
}

// 常规消息元素
type Elem struct {
	Text     string        `json:"text,omitempty"`
	Photo    *PhotoElem    `json:"photo,omitempty"`
	Sound    *SoundElem    `json:"sound,omitempty"`
	Video    *VideoElem    `json:"video,omitempty"`
	File     *FileElem     `json:"file,omitempty"`
	Mention  *MentionElem  `json:"mention,omitempty"`
	Location *LocationElem `json:"location,omitempty"`
	Quote    *QuoteElem    `json:"quote,omitempty"`
	Revoke   *RevokeElem   `json:"revoke,omitempty"`
	Custom   string        `json:"custom,omitempty"`
	ChatLog  *ChatLogElem  `json:"chat_log,omitempty"`
}

// 群创建成功提示
type TipChatCreated struct {
	Chat        *Chat   `json:"chat"`
	Creator     *User   `json:"creator"`
	InitMembers []*User `json:"init_members"`
}

// 新成员加入群提示
type TipMemberEnter struct {
	ChatId  string       `json:"chat_id"`
	Users   []*User      `json:"users"`
	Role    def.ChatRole `json:"role"`
	Inviter *TinyUser    `json:"inviter,omitempty"`
}

// 成员退出提示
type TipMemberQuit struct {
	ChatId string    `json:"chat_id"`
	User   *TinyUser `json:"user"`
}

// 群主转移提示
type TipOwnerTransfer struct {
	ChatId   string `json:"chat_id"`
	NewOwner *User  `json:"new_owner"`
	OldOwner *User  `json:"old_owner"`
}

// 移除成员提示
type TipMemberKicked struct {
	ChatId string    `json:"chat_id"`
	User   *TinyUser `json:"user"`
}

// 成员禁言/解禁提示
type TipMemberMuted struct {
	ChatId   string    `json:"chat_id"`  // 群ID
	Enable   bool      `json:"enable"`   // 开关: true禁言 false解禁
	User     *TinyUser `json:"user"`     // 被禁言的用户
	Duration int       `json:"duration"` // 持续秒数(0无限期)
}

// 群禁言/解禁提示
type TipChatMuted struct {
	ChatId   string `json:"chat_id"`  // 群ID
	Enable   bool   `json:"enable"`   // 开关: true禁言 false解禁
	Duration int    `json:"duration"` // 持续秒数(0无限期)
}

// 群修改成员角色提示
type TipEditRole struct {
	ChatId string    `json:"chat_id"` // 群ID
	User   *TinyUser `json:"user"`    // 群用户
	Role   int       `json:"role"`    // 成员角色: 0普通成员 1管理员
}

// 群解散提示
type TipDismissed struct {
	ChatId string `json:"chat_id"`
}

// 成为好友提示
// @Description 界面以文本消息的形式显示问候语.
// @Description 界面显示提示"以上是打招呼内容"
// @Description 界面显示"你已添加了<对方昵称>，现在可以开始聊天了。"
type TipBecomeFriends struct {
	Greets []*Greet `json:"greets"`  // 打招呼列表, 最新3条
	FromId string   `json:"from_id"` // 邀请者ID
	ToId   string   `json:"to_id"`   // 被邀请者ID
}

// 提示类消息
type Tip struct {
	ChatCreated   *TipChatCreated   `json:"chat_created,omitempty"`
	MemberEnter   *TipMemberEnter   `json:"member_enter,omitempty"`
	MemberQuit    *TipMemberQuit    `json:"member_quit,omitempty"`
	MemberKicked  *TipMemberKicked  `json:"member_kicked,omitempty"`
	OwnerTransfer *TipOwnerTransfer `json:"owner_transter,omitempty"`
	MemberMuted   *TipMemberMuted   `json:"member_muted,omitempty"`
	ChatMuted     *TipChatMuted     `json:"chat_muted,omitempty"`
	EditRole      *TipEditRole      `json:"edit_role,omitempty"`
	Dismissed     *TipDismissed     `json:"dismissed,omitempty"`
	BecomeFriends *TipBecomeFriends `json:"become_friends,omitempty"`
}

// 自己已读
// @Description 自己读了别人的消息, 多端同步用
type EvHasRead struct {
	PeerId string `json:"peer_id"` // 对话ID
	MaxId  int64  `json:"max_id"`  // 对话最新消息ID
	Unread int    `json:"unread"`  // 剩余未读数
}

// 正在输入
type EvTyping struct {
	PeerId string    `json:"peer_id"` // 对话ID
	User   *TinyUser `json:"user"`    // 对方用户信息
}

// 强制断线
// @Description 由CmdDisconnect触发, 凡是收到该事件的客户端都停止自动重连, 界面上提示断线原因
// @Description 原因类型
// @Description LOGOUT 			主动登出
// @Description OTHER_ONLINE    其他设备上线
// @Description BANNED          被封禁
// @Description DUPLICATED      重复连接
// @Description TOKEN_INVALID   TOKEN失效
type EvDisconnect struct {
	Reason string `json:"reason"` // 原因
}

// 主动退出群
type EvQuitChat struct {
	ChatId string `json:"chat_id"`
}

// 删除消息
type EvMsgDeleted struct {
	PeerId string  `json:"peer_id"`
	MsgId  []int64 `json:"msg_id"`
}

// 已读回执
type EvReceipt struct {
	PeerId string `json:"peer_id"` // 对话ID
	MaxId  int64  `json:"max_id"`  // 最新被标记为已读的消息ID(小于这个ID的消息都显示个对号)
}

// 修改成员昵称
type EvMemberName struct {
	ChatId string `json:"chat_id"` // 群ID
	UserId string `json:"user_id"` // 成员ID
	Name   string `json:"name"`    // 成员昵称
}

// 清空对话聊天记录(同时标记已读)
type EvDialogClear struct {
	PeerId string `json:"peer_id"`
	MaxId  int64  `json:"max_id"` // 当时的最大消息ID(清空<=max_id的聊天记录)
}

// 删除对话事件
type EvDialogDeleted struct {
	PeerId string `json:"peer_id"` // 对话ID
	Clear  bool   `json:"clear"`   // 是否清空聊天记录
	MaxId  int64  `json:"max_id"`  // 当时的最大消息ID(清空<=max_id的聊天记录)
}

// 置顶对话事件
type EvPinDialog struct {
	PeerId string `json:"peer_id"`
	Pinned bool   `json:"pinned"` // 置顶状态,true置顶,false取消置顶
}

// 更新免打扰设置事件
type EvUpdatePeerNotify struct {
	PeerId        string             `json:"peer_id"`
	NotifySetting *PeerNotifySetting `json:"notify_setting,omitempty"`
}

// 事件
type Event struct {
	Seq           int64               `json:"seq"` // 事件序号(在事件seq为0)
	HasRead       *EvHasRead          `json:"has_read,omitempty"`
	Typing        *EvTyping           `json:"typing,omitempty"`
	Disconnect    *EvDisconnect       `json:"disconnect,omitempty"`
	QuitChat      *EvQuitChat         `json:"quit_chat,omitempty"`
	MsgDeleted    *EvMsgDeleted       `json:"msg_deleted,omitempty"`
	Receipt       *EvReceipt          `json:"receipt,omitempty"`
	MemberName    *EvMemberName       `json:"member_name,omitempty"`
	DialogClear   *EvDialogClear      `json:"dialog_clear,omitempty"`
	DialogDeleted *EvDialogDeleted    `json:"dialog_deleted,omitempty"`
	PinDialog     *EvPinDialog        `json:"pin_dialog,omitempty"`
	PeerNotify    *EvUpdatePeerNotify `json:"peer_notify,omitempty"`
}

// 转发信息
type FwdHeader struct {
	FromId   string `json:"from_id"`   // 源头用户ID
	FromName string `json:"from_name"` // 源头用户昵称
	ChatId   string `json:"chat_id"`   // 源头群ID
	MsgId    int64  `json:"msg_id"`    // 源头消息ID
}

// 消息
type Message struct {
	Id        int64       `json:"id"` // 消息唯一ID(递增)
	Type      def.MsgType `json:"type"`
	FromId    string      `json:"from_id"`
	FromUser  *User       `json:"from_user,omitempty"`
	ToId      string      `json:"to_id"`
	Elem      *Elem       `json:"elem,omitempty"`
	Tip       *Tip        `json:"tip,omitempty"`
	Event     *Event      `json:"event,omitempty"`
	Revoked   bool        `json:"revoked"`              // 是否已撤消
	FwdHeader *FwdHeader  `json:"fwd_header,omitempty"` // 转发头
	CreatedAt int         `json:"created_at"`           // 消息发送时间
}

func (m *Message) GetPeerId(user_id string) string {
	if m.FromId == user_id {
		return m.ToId
	} else if m.ToId == user_id {
		return m.FromId
	}
	return m.ToId
}

func (m *Message) Blob() []byte {
	b, _ := json.Marshal(m)
	return b
}

// 对话
type Dialog struct {
	PeerId        string             `json:"peer_id"` // 以#开头的, 用来接收离线事件, 界面上不显示; #friend: 通讯录出红点用
	PeerUser      *User              `json:"peer_user,omitempty"`
	PeerChat      *Chat              `json:"peer_chat,omitempty"`
	Pinned        bool               `json:"pinned"`                   // 是否置顶
	Pts           int64              `json:"pts"`                      // 最新消息ID
	TopMessage    *Message           `json:"top_message"`              // 最新消息
	Unread        int                `json:"unread"`                   // 未读数
	ReceiptMaxId  int64              `json:"receipt_max_id"`           // 最大已读回执消息ID
	Seq           int64              `json:"seq"`                      // 最新事件ID
	NotifySetting *PeerNotifySetting `json:"notify_setting,omitempty"` // 免打扰设置
}

// 对话列表
type Dialogs struct {
	Dialogs []*Dialog `json:"dialogs"`
	Total   int       `json:"total"` // 总条数(该值只在第一页提供)
}

// 对话已读结果
type AffectedHistory struct {
	Pts    int64 `json:"pts"`    // 对话最新消息ID
	Unread int   `json:"unread"` // 剩余未读数
}

// 群成员
type Member struct {
	User  *User  `json:"user"`  // 成员用户信息
	Name  string `json:"name"`  // 成员昵称
	Muted bool   `json:"muted"` // 是否被禁言
	Role  int    `json:"role"`  // 角色(0普通成员 1管理员 2创建者)
}

// 问候语
type Greet struct {
	Text string `json:"text"` // 问候语文本
	Time int    `json:"time"` // 发送时间
}

// 好友申请记录
type FriendApply struct {
	Hash      string   `json:"hash"`                // 申请标识(唯一)
	FromId    string   `json:"from_id"`             // 申请者ID
	PeerUser  *User    `json:"peer_user,omitempty"` // 对方用户信息
	ToId      string   `json:"to_id"`               // 被邀请者ID
	Greets    []*Greet `json:"greets"`              // 问候语列表, 最新3条
	Name      string   `json:"name"`                // 申请者备注名称
	Status    int      `json:"status"`              // 状态(0等待验证 1已添加 2已过期)
	UpdatedAt int      `json:"updated_at"`          // 申请时间戳, 每次申请都会更新
}

// 好友
type Friend struct {
	UserId  string `json:"user_id"` // 用户ID
	Name    string `json:"name"`    // 备注昵称
	User    *User  `json:"user"`    // 好友用户信息
	Blocked bool   `json:"blocked"` // 是否被拉黑
	Letter  string `json:"letter"`  // 昵称首字母
}

// 用户免打扰设置
type PeerNotifySetting struct {
	Badge bool `json:"badge"` // 是否将未读数显示为小红点(true显示小红点 false显示未读数)
}

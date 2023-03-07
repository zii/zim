package def

// 客户端平台类型
type Platform int

func (p Platform) String() string {
	switch p {
	case IOS:
		return "IOS"
	case ANDROID:
		return "ANDROID"
	case WEB:
		return "WEB"
	case DESKTOP:
		return "DESKTOP"
	default:
		return "UNKNOWN"
	}
}

const (
	IOS     = Platform(1)
	ANDROID = Platform(2)
	WEB     = Platform(3)
	DESKTOP = Platform(4)
)

// redis消息pubsub channel名称
const MESSAGE_CHANNEL = "message:chan"

type IdType string

// UserId前缀类型
const (
	IdUser    = IdType("u") // 用户ID
	IdGroup   = IdType("g") // 普通群ID
	IdChannel = IdType("c") // 超级群ID
	IdSys     = IdType("#") // 系统ID
)

// 常用系统账号
const (
	IdFriend = "#friend"
)

func ToIdType(accid string) IdType {
	if accid == "" {
		return IdType("")
	}
	return IdType(accid[:1])
}

// 消息类型
type MsgType int

const (
	Msg        = MsgType(100) // 常规类消息
	MsgText    = MsgType(101) // 文本消息
	MsgPhoto   = MsgType(102) // 图片消息
	MsgSound   = MsgType(103) // 语音消息
	MsgVideo   = MsgType(104) // 视频消息
	MsgFile    = MsgType(105) // 文件消息
	MsgMention = MsgType(106) // @消息
	MsgQuote   = MsgType(107) // 引用消息
	MsgRevoke  = MsgType(108) // 撤销消息
	MsgChatlog = MsgType(109) // 聊天记录
	MsgCustom  = MsgType(190) // 自定义消息

	Tip              = MsgType(200) // 提示类消息
	TipChatCreated   = 201          // 群创建成功通知
	TipMemberEnter   = 202          // 新成员加入群通知
	TipMemberQuit    = 203          // 成员退出通知
	TipOwnerTransfer = 204          // 群主转移通知
	TipMemberKicked  = 205          // 移除成员通知
	TipMemberMuted   = 206          // 成员禁言/解禁
	TipChatMuted     = 207          // 群禁言/解禁
	TipEditRole      = 208          // 群修改成员角色
	TipDismissed     = MsgType(210) // 群解散通知
	TipBecomeFriends = MsgType(211) // 成为好友

	Event              = MsgType(300) // 事件类消息(并非消息, 只是一种行为)
	EvHasRead          = MsgType(301) // 自己已读别人的消息
	EvTyping           = MsgType(302) // 正在输入
	EvDisconnect       = MsgType(303) // 强制下线
	EvQuitChat         = MsgType(304) // 主动退群
	EvMsgDeleted       = MsgType(305) // 删除消息
	EvReceipt          = MsgType(306) // 已读回执
	EvMemberName       = MsgType(307) // 修改成员昵称
	EvDialogClear      = MsgType(308) // 清空对话聊天记录
	EvDialogDeleted    = MsgType(309) // 删除对话
	EvPinDialog        = MsgType(310) // 置顶对话
	EvUpdatePeerNotify = MsgType(311) // 更新免打扰设置
)

// Msg/Tip/Event
func (mt MsgType) Class() MsgType {
	return mt / 100 * 100
}

// 用户状态
const (
	UserOK     = 0 // 正常
	UserBanned = 1 // 封禁
)

// 群成员角色
type ChatRole int

const (
	RoleOwner  = ChatRole(2)
	RoleAdmin  = ChatRole(1)
	RoleMember = ChatRole(0)
)

// 内部命令操作类型
const (
	OpSend       = "send"
	OpDisconnect = "disconnect" // 踢在线用户连接
)

// 群类型
const (
	TypeGroup   = 1 // 普通群
	TypeChannel = 2 // 超级群
)

// 分页最大条数
const MaxPageSize = 100

// 最大保留问候语数量
const MaxGreetNum = 3

// 好友申请状态
const (
	FriendApplyWait   = 0
	FriendApplyAccept = 1
	FriendApplyExpire = 2
)

// 好友上限
const FriendLimit = 5000

// 分段计数器每段最大条数
const SegCounterLimit = 10000

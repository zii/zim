package def

// 是否启用多数据中心模式
// 多中心: 用户ID/群ID/消息ID用sonyflake算法生成, ID很长
// 单中心: 用户ID/群ID/消息ID用redis计数器生成, ID很短
var UseMultiDC bool

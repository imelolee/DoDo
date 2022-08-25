package config

import "time"

// Secret 密钥
var Secret = "DoDo"

// OneDayOfHours 时间
var OneDayOfHours = 60 * 60 * 24
var OneMinute = 60 * 1
var OneMonth = 60 * 60 * 24 * 30
var OneYear = 365 * 60 * 60 * 24
var ExpireTime = time.Hour * 48 // 设置Redis数据热度消散时间。

const ValidComment = 0   //评论状态：有效
const InvalidComment = 1 //评论状态：取消
const DateTime = "2006-01-02 15:04:05"

const DefaultRedisValue = -1 //redis中key对应的预设值，防脏读

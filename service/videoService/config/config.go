package config

import (
	"go-micro.dev/v4/client"
	"time"
)

// Secret 密钥
var Secret = "DoDo"

// OneDayOfHours 时间
var OneDayOfHours = 60 * 60 * 24
var OneMinute = 60 * 1
var OneMonth = 60 * 60 * 24 * 30
var OneYear = 365 * 60 * 60 * 24
var ExpireTime = time.Hour * 48 // 设置Redis数据热度消散时间。

// VideoCount 每次获取视频流的数量
const VideoCount = 5

var Opts client.CallOption = func(o *client.CallOptions) {
	o.RequestTimeout = time.Second * 30
	o.DialTimeout = time.Second * 30
}

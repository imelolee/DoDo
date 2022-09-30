# DoDo
[![go](https://img.shields.io/badge/go-1.8-blue)](https://github.com/golang/go)
[![gin](https://img.shields.io/badge/gin-v1.8-blue)](https://github.com/gin-gonic/gin/releases)
[![consul](https://img.shields.io/badge/consul-v1.10-blue)](https://github.com/hashicorp/consul)
[![go-micro](https://img.shields.io/badge/go--micro-v4-blue)](https://github.com/go-micro/go-micro)



## 项目结构
```
DoDo 
├── config -- 配置类
├── controller -- 控制器相关
├── middleware -- 中间件相关
├── model -- 数据库模型相关
├── proto -- protobuf相关
├── utils -- 工具类
└── service -- 微服务模块
    ├── commentService -- 评论服务相关
    ├── followService -- 关注服务相关
    ├── likeService -- 点赞服务相关
    ├── userService -- 用户服务相关
    └── videoService -- 视频服务相关
```

## 接口文档
[https://www.apifox.cn/apidoc/shared-8cc50618-0da6-4d5e-a398-76f3b8f766c5/api-18345145](https://www.apifox.cn/apidoc/shared-8cc50618-0da6-4d5e-a398-76f3b8f766c5/api-18345145)

## 步骤

#### 启动Consul服务发现
```
consul agent -dev
```
访问 [http://localhost:8500](http://localhost:8500) 监控服务状态

#### 启动各项微服务
```
cd service/commentService
go run main.go

cd service/followService
go run main.go

cd service/likeService
go run main.go

cd service/userService
go run main.go

cd service/videoService
go run main.go

```
#### 启动web服务
```
go run main.go
```

package util

import (
	"bytes"
	"context"
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	// 初始化客户端
	consulReg := consul.NewRegistry()

	return micro.NewService(micro.Registry(consulReg))

}

// GetMicroClient 初始化客户端
func GetMicroClient() client.Client {
	consulReg := consul.NewRegistry()
	microService := micro.NewService(
		micro.Registry(consulReg),
	)
	return microService.Client()
}

func getRandstring(length int) string {
	if length < 1 {
		return ""
	}
	char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charArr := strings.Split(char, "")
	charlen := len(charArr)
	ran := rand.New(rand.NewSource(time.Now().Unix()))
	var rchar string = ""
	for i := 1; i <= length; i++ {
		rchar = rchar + charArr[ran.Intn(charlen)]
	}
	return rchar
}

// RandFileName 随机文件名
func RandFileName(fileName string) string {
	randStr := getRandstring(16)
	return randStr + filepath.Ext(fileName)
}

// UpLoadQiniu 七牛云上传
func UpLoadQiniu(file []byte, fileName string, fileSize int64) (key string, err error) {
	// 生成随即文件名
	key = RandFileName(fileName)

	putPolicy := storage.PutPolicy{
		Scope: Bucket,
	}
	// 获取上传凭证
	mac := qbox.NewMac(AccessKey, SecretKey)
	upToken := putPolicy.UploadToken(mac)

	// 存储桶配置
	cfg := storage.Config{
		Zone:          &storage.ZoneHuabei,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}

	putExtra := storage.PutExtra{}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	fileReader := bytes.NewReader(file)
	err = formUploader.Put(context.Background(), &ret, upToken, key, fileReader, fileSize, &putExtra)

	return key, err
}

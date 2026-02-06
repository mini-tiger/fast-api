package aliOss

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/mini-tiger/fast-api/config"
	"github.com/mini-tiger/fast-api/core"
)

type ClientType struct {
	cdnHost             string
	bucketName          string
	client              *oss.Client
	useInternalEndpoint bool
}

func New() *ClientType {
	ossConfig := config.GetInstance().Section("aliOss")
	return &ClientType{
		bucketName: ossConfig.Key("bucketName").Value(),
		cdnHost:    ossConfig.Key("cdnHost").Value(),
	}
}

func (c *ClientType) GetClient() *oss.Client {
	if c.client != nil {
		return c.client
	}
	ossConfig := config.GetInstance().Section("aliOss")
	endpoint := ossConfig.Key("endpoint").Value()
	region := ossConfig.Key("region").Value()
	accessKeyID := ossConfig.Key("accessKeyId").Value()
	accessKeySecret := ossConfig.Key("accessKeySecret").Value()

	_ = os.Setenv("OSS_ACCESS_KEY_ID", accessKeyID)
	_ = os.Setenv("OSS_ACCESS_KEY_SECRET", accessKeySecret)

	// 方式二：同时填写Region和Endpoint
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region).     // 填写Bucket所在地域
		WithEndpoint(endpoint). // 填写Bucket所在地域对应的公网Endpoint
		WithUseInternalEndpoint(c.useInternalEndpoint)
	c.client = oss.NewClient(cfg)
	return c.client
}

func (c *ClientType) SetUseInternalEndpoint(useInternalEndpoint bool) *ClientType {
	c.useInternalEndpoint = useInternalEndpoint
	return c
}

// UploadFile 上传文件
func (c *ClientType) UploadFile(localFilePath, dir, fileName string, metadata map[string]string) (string, error) {
	// 服务端上传用内网
	c.SetUseInternalEndpoint(true)
	// 开发环境用外网
	if core.Mode == core.Dev {
		c.SetUseInternalEndpoint(false)
	}
	// 获取本地文件名
	serverName := config.GetInstance().Section("core").Key("serverName").Value()
	cdnHost := config.GetInstance().Section("aliOss").Key("cdnHost").Value()

	date := time.Now().Format("200601/02")
	unixMicro := time.Now().UnixMicro()

	// 组装上传后的oss地址
	filePath := fmt.Sprintf("%s-%s/%s/%s/%d/%s", serverName, core.Mode, dir, date, unixMicro, fileName)

	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(c.bucketName),    // 存储空间名称
		Key:          oss.Ptr(filePath),        // 对象名称
		StorageClass: oss.StorageClassStandard, // 指定对象的存储类型为标准存储
		Acl:          oss.ObjectACLPublicRead,  // 指定对象的访问权限为私有访问
		Metadata:     metadata,
	}
	// 上传
	_, err := c.GetClient().PutObjectFromFile(context.Background(), putRequest, localFilePath)
	if err != nil {
		return "", err
	}
	// 返回新地址
	return fmt.Sprintf("%s/%s", cdnHost, filePath), nil
}

type GetPreUploadFileUrlResultType struct {
	TokenUrl string
	FileUrl  string
	Error    error
}

// GetPreUploadFileUrl 前端上传文件获取url
func (c *ClientType) GetPreUploadFileUrl(dir, fileName string, metadata map[string]string) GetPreUploadFileUrlResultType {
	// 前端上传需要用外网
	c.SetUseInternalEndpoint(false)

	// 获取本地文件名
	serverName := config.GetInstance().Section("core").Key("serverName").Value()
	cdnHost := config.GetInstance().Section("aliOss").Key("cdnHost").Value()

	date := time.Now().Format("200601/02")
	unixMicro := time.Now().UnixMicro()

	// 组装上传后的oss地址
	filePath := fmt.Sprintf("%s-%s/%s/%s/%d/%s", serverName, core.Mode, dir, date, unixMicro, fileName)

	// 生成PutObject的预签名URL
	result, err := c.GetClient().Presign(context.TODO(), &oss.PutObjectRequest{
		Bucket:       oss.Ptr(c.bucketName),
		Key:          oss.Ptr(filePath),
		ContentType:  oss.Ptr("text/plain;charset=utf8"), // 请确保在服务端生成该签名URL时设置的ContentType与在使用URL时设置的ContentType一致
		StorageClass: oss.StorageClassStandard,           // 请确保在服务端生成该签名URL时设置的StorageClass与在使用URL时设置的StorageClass一致
		//Metadata:     map[string]string{"key1": "value1", "key2": "value2"}, // 请确保在服务端生成该签名URL时设置的Metadata与在使用URL时设置的Metadata一致
	},
		oss.PresignExpires(10*time.Minute),
	)

	if err != nil {
		return GetPreUploadFileUrlResultType{
			TokenUrl: "",
			FileUrl:  "",
			Error:    err,
		}
	}

	return GetPreUploadFileUrlResultType{
		TokenUrl: result.URL,
		FileUrl:  fmt.Sprintf("%s/%s", cdnHost, filePath),
		Error:    nil,
	}
}

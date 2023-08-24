package api

import (
	"context"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	console "github.com/alibabacloud-go/tea-console/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"mxshop-api/user-web/forms"
	"mxshop-api/user-web/global"
	"net/http"
	"strings"
	"time"
)

func GenerateSmsCode(witdh int) string {
	//生成width长度的短信验证码

	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < witdh; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

//func CreateClient() (*dysmsapi20170525.Client, error) {
//	config := &openapi.Config{
//		AccessKeyId:     tea.String(global.ServerConfig.AliSmsInfo.ApiKey),
//		AccessKeySecret: tea.String(global.ServerConfig.AliSmsInfo.ApiSecrect),
//	}
//	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
//	return dysmsapi20170525.NewClient(config)
//}

func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

func SendSms(ctx *gin.Context) {
	sendSmsForm := forms.SendSmsForm{}
	if err := ctx.ShouldBind(&sendSmsForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}
	client, err := CreateClient(tea.String(global.ServerConfig.AliSmsInfo.ApiKey), tea.String(global.ServerConfig.AliSmsInfo.ApiSecrect))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	smsCode := GenerateSmsCode(6)
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String("go商城"),
		TemplateCode:  tea.String("SMS_462615325"),
		PhoneNumbers:  tea.String(sendSmsForm.Mobile),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", smsCode)),
	}
	runtime := &util.RuntimeOptions{}
	resp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	console.Log(util.ToJSONString(resp))
	// 将验证码保存起来 - redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	err = rdb.Set(context.Background(), sendSmsForm.Mobile, smsCode, time.Duration(global.ServerConfig.RedisInfo.Expire)*time.Second).Err()
	if err != nil {
		fmt.Println("Redis set error:", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "发送成功",
	})
}

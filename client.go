package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

//AgentClient 是用于主动调用微信API的Agent客户端
type AgentClient struct {
	CorpId  string
	AgentId int
	Secret  string

	AccessToken          string
	AccessTokenExpiresAt time.Time
}

//NewAgentClientFromEnv 从环境变量中获取配置信息并创建一个微信企业应用的客户端，如果未获取到配置信息则打印警告信息
func NewAgentClientFromEnv() *AgentClient {
	corpId := os.Getenv("WECHAT_CORP_ID")
	secret := os.Getenv("WECHAT_SECRET")
	if corpId == "" || secret == "" {
		fmt.Println("WARN: WECHAT_CORP_ID or WECHAT_SECRET not be set")
	}

	agentId, _ := strconv.Atoi(os.Getenv("WECHAT_AGENT_ID"))
	return NewAgentClient(corpId, agentId, secret)
}

//NewAgentClient 创建新的微信企业应用客户端
func NewAgentClient(corpId string, agentId int, secret string) *AgentClient {
	return &AgentClient{
		CorpId:  corpId,
		AgentId: agentId,
		Secret:  secret,
	}
}

//CommonResponse 是微信API公用返回结构
type CommonResponse struct {
	ErrCode int
	ErrMsg  string
}

//WechatApiUrl 保存微信企业号API服务器地址
const WechatApiUrl = "https://qyapi.weixin.qq.com/cgi-bin"

func (cli *AgentClient) requestWithToken(method, path string, query url.Values, reqData interface{}, respInfo interface{}) error {
	token, err := cli.GetAccessTokenFromCache()
	if err != nil {
		return fmt.Errorf("获取Token错误: %s", err)
	}

	if query == nil {
		query = url.Values{}
	}
	query.Set("access_token", token)

	buf, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("编码参数错误: %s", err)
	}

	req, err := http.NewRequest(method, WechatApiUrl+path+"?"+query.Encode(), bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("构造请求错误: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求错误: %s", err)
	}

	defer resp.Body.Close()

	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应错误: %s", err)
	}

	var info CommonResponse
	err = json.Unmarshal(buf, &info)
	if err != nil {
		return fmt.Errorf("解析响应错误: %s", err)
	}

	if info.ErrCode != 0 {
		return fmt.Errorf("API错误: [%d]%s", info.ErrCode, info.ErrMsg)
	}

	err = json.Unmarshal(buf, &respInfo)
	if err != nil {
		return fmt.Errorf("解析响应错误: %s", err)
	}

	return nil
}

package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/carmel/warp-cli/util"
)

const (
	TeamAuthDomain = "cloudflareaccess.com"
)

// TeamDevice 表示 Zero Trust 设备
type TeamDevice struct {
	ID     string `json:"id"`
	Token  string `json:"token"`
	Config struct {
		Peers []struct {
			PublicKey string `json:"public_key"`
			Endpoint  struct {
				Host string `json:"host"`
			} `json:"endpoint"`
		} `json:"peers"`
		Interface struct {
			Addresses struct {
				V4 string `json:"v4"`
				V6 string `json:"v6"`
			} `json:"addresses"`
		} `json:"interface"`
	} `json:"config"`
	Account struct {
		ID          string `json:"id"`
		AccountType string `json:"account_type"`
		License     string `json:"license"`
	} `json:"account"`
}

// TeamRegisterRequest 注册请求
type TeamRegisterRequest struct {
	Key       string `json:"key"`
	InstallID string `json:"install_id"`
	FcmToken  string `json:"fcm_token"`
	Tos       string `json:"tos"`
	Type      string `json:"type"`
	Model     string `json:"model"`
	Locale    string `json:"locale"`
}

// RegisterTeamDevice 使用 token 注册 Zero Trust 设备
func RegisterTeamDevice(teamName, token string, publicKey *util.Key) (*TeamDevice, error) {
	// Zero Trust 注册实际上是两步过程：
	// 1. 先在标准 WARP API 注册设备
	// 2. 使用 JWT token 将设备关联到 Zero Trust 组织

	// 第一步：标准 WARP 注册
	url := fmt.Sprintf("%s/%s/reg", ApiUrl, ApiVersion)

	// 构建请求体
	reqBody := TeamRegisterRequest{
		Key:       publicKey.String(),
		InstallID: "",
		FcmToken:  "",
		Tos:       util.GetTimestamp(),
		Type:      "Linux",
		Model:     "PC",
		Locale:    "en_US",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置 headers - 使用 JWT token 进行认证
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", DefaultHeaders["User-Agent"])
	req.Header.Set("CF-Client-Version", DefaultHeaders["CF-Client-Version"])

	// 发送请求
	client := &http.Client{
		Transport: DefaultTransport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var device TeamDevice
	if err := json.Unmarshal(body, &device); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	return &device, nil
}

// RegisterTeamDeviceWithServiceToken 使用 Service Token 注册
func RegisterTeamDeviceWithServiceToken(
	teamName, clientID, clientSecret string,
	publicKey *util.Key,
) (*TeamDevice, error) {
	url := fmt.Sprintf("https://%s.%s/warp", teamName, TeamAuthDomain)

	// 构建请求体
	reqBody := TeamRegisterRequest{
		Key:       publicKey.String(),
		InstallID: "",
		FcmToken:  "",
		Tos:       util.GetTimestamp(),
		Type:      "Linux",
		Model:     "PC",
		Locale:    "en_US",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置 headers（Service Token 认证）
	req.Header.Set("CF-Access-Client-Id", clientID)
	req.Header.Set("CF-Access-Client-Secret", clientSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", DefaultHeaders["User-Agent"])
	req.Header.Set("CF-Client-Version", DefaultHeaders["CF-Client-Version"])

	// 发送请求
	client := &http.Client{
		Transport: DefaultTransport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var device TeamDevice
	if err := json.Unmarshal(body, &device); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	return &device, nil
}

// GetTeamDevice 获取 Zero Trust 设备信息
func GetTeamDevice(teamName, deviceID, accessToken string) (*TeamDevice, error) {
	url := fmt.Sprintf("https://%s.%s/warp/%s", teamName, TeamAuthDomain, deviceID)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", DefaultHeaders["User-Agent"])

	client := &http.Client{
		Transport: DefaultTransport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var device TeamDevice
	if err := json.Unmarshal(body, &device); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	return &device, nil
}

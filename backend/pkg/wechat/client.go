package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"project-management/internal/config"
)

type WeChatClient struct {
	AppID     string
	AppSecret string
}

type QRCodeResponse struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	URL           string `json:"url"`
}

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

type UserInfoResponse struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string `json:"unionid"`
}

func NewWeChatClient() *WeChatClient {
	return &WeChatClient{
		AppID:     config.AppConfig.WeChat.AppID,
		AppSecret: config.AppConfig.WeChat.AppSecret,
	}
}

// GetQRCode 获取微信登录二维码
func (c *WeChatClient) GetQRCode() (*QRCodeResponse, error) {
	// 这里应该调用微信开放平台API获取二维码
	// 由于需要实际的AppID和AppSecret，这里先返回模拟数据
	// 实际实现需要调用: https://api.weixin.qq.com/cgi-bin/qrcode/create
	
	return &QRCodeResponse{
		Ticket:        "mock_ticket_" + fmt.Sprintf("%d", time.Now().Unix()),
		ExpireSeconds: 300,
		URL:           "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=mock_ticket",
	}, nil
}

// GetQRCodeURL 获取二维码图片URL
func (c *WeChatClient) GetQRCodeURL(ticket string) string {
	return fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", ticket)
}

// GetAccessToken 通过code获取access_token
func (c *WeChatClient) GetAccessToken(code string) (*AccessTokenResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.AppID,
		c.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.AccessToken == "" {
		return nil, fmt.Errorf("failed to get access token: %s", string(body))
	}

	return &result, nil
}

// GetUserInfo 获取用户信息
func (c *WeChatClient) GetUserInfo(accessToken, openID string) (*UserInfoResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken,
		openID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result UserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.OpenID == "" {
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	return &result, nil
}


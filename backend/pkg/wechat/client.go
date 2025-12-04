package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"project-management/internal/config"
)

type WeChatClient struct {
	AppID       string
	AppSecret   string
	AccountType string // "official_account" 或 "open_platform"
	Scope       string // "snsapi_base" 或 "snsapi_userinfo" (仅公众号使用)
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
	accountType := config.AppConfig.WeChat.AccountType
	if accountType == "" {
		accountType = "open_platform" // 默认使用开放平台
	}
	scope := config.AppConfig.WeChat.Scope
	if scope == "" {
		scope = "snsapi_userinfo" // 默认需要用户确认
	}
	return &WeChatClient{
		AppID:       config.AppConfig.WeChat.AppID,
		AppSecret:   config.AppConfig.WeChat.AppSecret,
		AccountType: accountType,
		Scope:       scope,
	}
}

// GetQRCode 获取微信登录授权URL
// 根据 AccountType 生成不同的授权URL：
// - open_platform: 开放平台网站应用扫码登录（qrconnect）
//   - 用户扫码后，在微信中确认授权，然后跳转到回调地址
// - official_account: 公众号网页授权（oauth2/authorize）
//   - 可以生成二维码，用户扫码后会在微信内打开授权页面
//   - 用户确认授权后，跳转到回调地址
// 注意：两种方式都可以生成二维码，前端会将授权URL转换为二维码图片
// customState: 可选的自定义state参数，如果为空则使用时间戳
func (c *WeChatClient) GetQRCode(redirectURI string, customState ...string) (*QRCodeResponse, error) {
	if c.AppID == "" {
		return nil, fmt.Errorf("微信AppID未配置")
	}

	var state string
	if len(customState) > 0 && customState[0] != "" {
		state = customState[0]
	} else {
		state = fmt.Sprintf("%d", time.Now().Unix())
	}
	encodedRedirectURI := url.QueryEscape(redirectURI)
	
	var authURL string
	
	if c.AccountType == "official_account" {
		// 公众号网页授权URL
		// 格式：https://open.weixin.qq.com/connect/oauth2/authorize?appid=APPID&redirect_uri=REDIRECT_URI&response_type=code&scope=SCOPE&state=STATE#wechat_redirect
		// 注意：redirect_uri 必须进行URL编码，且必须在公众号后台配置的"网页授权域名"中
		// 说明：可以生成二维码，用户扫码后会在微信内打开授权页面，确认后跳转到回调地址
		scope := c.Scope
		if scope == "" {
			scope = "snsapi_userinfo"
		}
		authURL = fmt.Sprintf(
			"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
			c.AppID,
			encodedRedirectURI,
			scope,
			state,
		)
	} else {
		// 微信开放平台网站应用扫码登录授权URL
		// 格式：https://open.weixin.qq.com/connect/qrconnect?appid=APPID&redirect_uri=REDIRECT_URI&response_type=code&scope=snsapi_login&state=STATE#wechat_redirect
		// 注意：redirect_uri 必须进行URL编码，且必须在微信开放平台配置的授权回调域名中
		authURL = fmt.Sprintf(
			"https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
			c.AppID,
			encodedRedirectURI,
			state,
		)
	}

	// 返回授权URL，前端需要将其转换为二维码图片
	// 两种方式都支持生成二维码：
	// - 开放平台：扫码后在微信中确认授权
	// - 公众号：扫码后在微信内打开授权页面
	// 使用 state 作为 ticket，用于后续验证
	return &QRCodeResponse{
		Ticket:        state,
		ExpireSeconds: 600, // 微信授权链接有效期10分钟
		URL:           authURL,
	}, nil
}

// GetQRCodeURL 获取二维码图片URL
// 注意：这里返回的是授权URL，前端需要使用二维码生成库将其转换为二维码图片
func (c *WeChatClient) GetQRCodeURL(ticket string) string {
	// 如果 ticket 是授权URL，直接返回
	if len(ticket) > 20 && ticket[:4] == "http" {
		return ticket
	}
	// 否则返回空，因为真实的二维码需要通过授权URL生成
	return ""
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

	// 先检查是否是错误响应
	var errorResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.ErrCode != 0 {
		// 根据错误码提供更详细的错误信息
		switch errorResp.ErrCode {
		case 40013:
			// invalid appsecret
			return nil, fmt.Errorf("AppSecret无效。请检查：1) AppID和AppSecret是否匹配（必须来自同一个微信应用）；2) AppSecret是否正确（注意不要有多余的空格）；3) 如果重置过AppSecret，请使用最新的AppSecret；4) 确认使用的是公众号AppSecret还是开放平台AppSecret: %s", errorResp.ErrMsg)
		case 40029:
			return nil, fmt.Errorf("code已过期或已使用，请重新扫码: %s", errorResp.ErrMsg)
		case 40163:
			return nil, fmt.Errorf("code已被使用，请重新扫码: %s", errorResp.ErrMsg)
		case 40125:
			return nil, fmt.Errorf("scope参数错误或没有scope权限。请检查：1) 授权回调域名是否正确配置（只填写域名，不包含协议和端口）；2) 是否使用了正确的AppID类型（开放平台网站应用使用 open_platform，公众号使用 official_account）；3) 是否已申请微信登录接口权限；4) AppID和AppSecret是否匹配: %s", errorResp.ErrMsg)
		default:
			return nil, fmt.Errorf("微信接口错误 (errcode: %d): %s", errorResp.ErrCode, errorResp.ErrMsg)
		}
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

// SetAppID 设置AppID
func (c *WeChatClient) SetAppID(appID string) {
	c.AppID = appID
}

// SetAppSecret 设置AppSecret
func (c *WeChatClient) SetAppSecret(appSecret string) {
	c.AppSecret = appSecret
}

// SetAccountType 设置AccountType
func (c *WeChatClient) SetAccountType(accountType string) {
	c.AccountType = accountType
}

// SetScope 设置Scope
func (c *WeChatClient) SetScope(scope string) {
	c.Scope = scope
}

// GetAppID 获取AppID
func (c *WeChatClient) GetAppID() string {
	return c.AppID
}

// GetAppSecret 获取AppSecret
func (c *WeChatClient) GetAppSecret() string {
	return c.AppSecret
}

// GetAccountType 获取AccountType
func (c *WeChatClient) GetAccountType() string {
	return c.AccountType
}

// GetScope 获取Scope
func (c *WeChatClient) GetScope() string {
	return c.Scope
}

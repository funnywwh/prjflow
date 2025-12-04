package mocks

import (
	"fmt"

	"project-management/pkg/wechat"
)

// MockWeChatClient 微信客户端Mock实现
type MockWeChatClient struct {
	// GetAccessToken的返回值配置
	AccessTokenResponse *wechat.AccessTokenResponse
	AccessTokenError    error

	// GetUserInfo的返回值配置
	UserInfoResponse *wechat.UserInfoResponse
	UserInfoError    error

	// GetQRCode的返回值配置
	QRCodeResponse *wechat.QRCodeResponse
	QRCodeError    error

	// 记录调用次数
	GetAccessTokenCallCount int
	GetUserInfoCallCount    int
	GetQRCodeCallCount      int

	// 记录调用参数
	GetAccessTokenCodes []string
	GetUserInfoParams   []struct {
		AccessToken string
		OpenID      string
	}
	GetQRCodeParams []struct {
		RedirectURI string
		CustomState []string
	}

	// 配置字段（用于getter/setter）
	AppID       string
	AppSecret   string
	AccountType string
	Scope       string
}

// NewMockWeChatClient 创建新的MockWeChatClient
func NewMockWeChatClient() *MockWeChatClient {
	return &MockWeChatClient{
		GetAccessTokenCodes: make([]string, 0),
		GetUserInfoParams:   make([]struct{ AccessToken, OpenID string }, 0),
		GetQRCodeParams:     make([]struct{ RedirectURI string; CustomState []string }, 0),
	}
}

// GetAccessToken 实现WeChatClientInterface接口
func (m *MockWeChatClient) GetAccessToken(code string) (*wechat.AccessTokenResponse, error) {
	m.GetAccessTokenCallCount++
	m.GetAccessTokenCodes = append(m.GetAccessTokenCodes, code)

	if m.AccessTokenError != nil {
		return nil, m.AccessTokenError
	}

	if m.AccessTokenResponse == nil {
		return nil, fmt.Errorf("AccessTokenResponse not configured")
	}

	return m.AccessTokenResponse, nil
}

// GetUserInfo 实现WeChatClientInterface接口
func (m *MockWeChatClient) GetUserInfo(accessToken, openID string) (*wechat.UserInfoResponse, error) {
	m.GetUserInfoCallCount++
	m.GetUserInfoParams = append(m.GetUserInfoParams, struct {
		AccessToken string
		OpenID      string
	}{AccessToken: accessToken, OpenID: openID})

	if m.UserInfoError != nil {
		return nil, m.UserInfoError
	}

	if m.UserInfoResponse == nil {
		return nil, fmt.Errorf("UserInfoResponse not configured")
	}

	return m.UserInfoResponse, nil
}

// GetQRCode 实现WeChatClientInterface接口
func (m *MockWeChatClient) GetQRCode(redirectURI string, customState ...string) (*wechat.QRCodeResponse, error) {
	m.GetQRCodeCallCount++
	m.GetQRCodeParams = append(m.GetQRCodeParams, struct {
		RedirectURI string
		CustomState []string
	}{RedirectURI: redirectURI, CustomState: customState})

	if m.QRCodeError != nil {
		return nil, m.QRCodeError
	}

	if m.QRCodeResponse == nil {
		return nil, fmt.Errorf("QRCodeResponse not configured")
	}

	return m.QRCodeResponse, nil
}

// SetAppID 设置AppID
func (m *MockWeChatClient) SetAppID(appID string) {
	m.AppID = appID
}

// SetAppSecret 设置AppSecret
func (m *MockWeChatClient) SetAppSecret(appSecret string) {
	m.AppSecret = appSecret
}

// SetAccountType 设置AccountType
func (m *MockWeChatClient) SetAccountType(accountType string) {
	m.AccountType = accountType
}

// SetScope 设置Scope
func (m *MockWeChatClient) SetScope(scope string) {
	m.Scope = scope
}

// GetAppID 获取AppID
func (m *MockWeChatClient) GetAppID() string {
	return m.AppID
}

// GetAppSecret 获取AppSecret
func (m *MockWeChatClient) GetAppSecret() string {
	return m.AppSecret
}

// GetAccountType 获取AccountType
func (m *MockWeChatClient) GetAccountType() string {
	return m.AccountType
}

// GetScope 获取Scope
func (m *MockWeChatClient) GetScope() string {
	return m.Scope
}

// Reset 重置Mock状态
func (m *MockWeChatClient) Reset() {
	m.AccessTokenResponse = nil
	m.AccessTokenError = nil
	m.UserInfoResponse = nil
	m.UserInfoError = nil
	m.QRCodeResponse = nil
	m.QRCodeError = nil
	m.GetAccessTokenCallCount = 0
	m.GetUserInfoCallCount = 0
	m.GetQRCodeCallCount = 0
	m.GetAccessTokenCodes = make([]string, 0)
	m.GetUserInfoParams = make([]struct{ AccessToken, OpenID string }, 0)
	m.GetQRCodeParams = make([]struct{ RedirectURI string; CustomState []string }, 0)
	m.AppID = ""
	m.AppSecret = ""
	m.AccountType = ""
	m.Scope = ""
}


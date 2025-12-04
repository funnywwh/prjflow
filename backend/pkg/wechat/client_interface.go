package wechat

// WeChatClientInterface 微信客户端接口
// 用于依赖注入和测试mock
type WeChatClientInterface interface {
	// GetAccessToken 通过code获取access_token
	GetAccessToken(code string) (*AccessTokenResponse, error)

	// GetUserInfo 获取用户信息
	GetUserInfo(accessToken, openID string) (*UserInfoResponse, error)

	// GetQRCode 获取微信登录授权URL
	GetQRCode(redirectURI string, customState ...string) (*QRCodeResponse, error)

	// SetAppID 设置AppID（用于配置）
	SetAppID(appID string)

	// SetAppSecret 设置AppSecret（用于配置）
	SetAppSecret(appSecret string)

	// SetAccountType 设置AccountType（用于配置）
	SetAccountType(accountType string)

	// SetScope 设置Scope（用于配置）
	SetScope(scope string)

	// GetAppID 获取AppID（用于验证）
	GetAppID() string

	// GetAppSecret 获取AppSecret（用于验证）
	GetAppSecret() string

	// GetAccountType 获取AccountType
	GetAccountType() string

	// GetScope 获取Scope
	GetScope() string
}

// 确保WeChatClient实现了WeChatClientInterface接口
var _ WeChatClientInterface = (*WeChatClient)(nil)


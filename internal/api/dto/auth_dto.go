package dto

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string  `json:"username" binding:"required,min=1,max=255"`
	Password        string  `json:"password" binding:"required,min=6,max=255"`
	Avatar          *string `json:"avatar" binding:"omitempty,max=500"`
	BackgroundImage *string `json:"background_image" binding:"omitempty,max=500"`
	UserRole        string  `json:"user_role" binding:"omitempty,oneof=user admin"`
}

// TokenData 登录成功返回的 Token 信息
type TokenData struct {
	Token     string   `json:"token"`
	TokenType string   `json:"token_type"`
	ExpiresIn int      `json:"expires_in"`
	User      UserInfo `json:"user"`
}

// UserInfo 用户公开信息（不含密码）
type UserInfo struct {
	ID              int64   `json:"id"`
	Username        string  `json:"user_name"`
	Avatar          *string `json:"avatar"`
	BackgroundImage *string `json:"background_image"`
	UserRole        string  `json:"user_role"`
	FollowCount     int64   `json:"follow_count"`
	FollowerCount   int64   `json:"follower_count"`
}

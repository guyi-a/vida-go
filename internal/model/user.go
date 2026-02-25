package model

// User 用户模型
type User struct {
	ID              int64   `gorm:"primaryKey;autoIncrement;comment:用户标识" json:"id"`
	UserName        string  `gorm:"size:255;not null;uniqueIndex;comment:用户名" json:"user_name"`
	Password        string  `gorm:"size:255;not null;comment:密码" json:"-"` // json:"-" 序列化时忽略密码
	FollowCount     int64   `gorm:"not null;default:0;comment:关注其他用户个数" json:"follow_count"`
	FollowerCount   int64   `gorm:"not null;default:0;comment:粉丝个数" json:"follower_count"`
	TotalFavorited  int64   `gorm:"not null;default:0;comment:用户被喜欢的视频数量" json:"total_favorited"`
	FavoriteCount   int64   `gorm:"not null;default:0;comment:用户喜欢的视频数量" json:"favorite_count"`
	Avatar          *string `gorm:"size:500;comment:用户头像" json:"avatar"`
	BackgroundImage *string `gorm:"size:500;comment:主页背景" json:"background_image"`
	UserRole        string  `gorm:"size:256;not null;default:'user';comment:用户角色" json:"user_role"`
	IsDelete        int64   `gorm:"not null;default:0;comment:删除标识" json:"-"`

	// 关联关系
	Videos    []Video    `gorm:"foreignKey:AuthorID" json:"videos,omitempty"`
	Favorites []Favorite `gorm:"foreignKey:UserID" json:"favorites,omitempty"`
	Comments  []Comment  `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

func (User) TableName() string {
	return "users"
}

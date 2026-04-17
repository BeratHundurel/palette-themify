package model

import "time"

type Color struct {
	Hex string `json:"hex"`
}

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:255;not null;index"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	GoogleID     string    `json:"googleId" gorm:"size:255;index"`
	AvatarURL    string    `json:"avatarUrl" gorm:"size:512"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Palettes     []Palette `json:"palettes" gorm:"foreignKey:UserID"`
	Themes       []Theme   `json:"themes" gorm:"foreignKey:UserID"`
}

type Palette struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    *uint      `json:"userId" gorm:"index"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name      string     `json:"name" gorm:"size:255;not null"`
	JsonData  string     `json:"jsonData" gorm:"type:jsonb;not null"`
	IsSystem  bool       `json:"isSystem" gorm:"default:false"`
	IsShared  bool       `json:"isShared" gorm:"default:false;index"`
	SharedAt  *time.Time `json:"sharedAt" gorm:"index"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type Theme struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	UserID     *uint      `json:"userId" gorm:"index;uniqueIndex:idx_user_theme_signature"`
	Name       string     `json:"name" gorm:"size:255;not null"`
	EditorType string     `json:"editorType" gorm:"size:20;uniqueIndex:idx_user_theme_signature"`
	Signature  string     `json:"signature" gorm:"size:128;uniqueIndex:idx_user_theme_signature"`
	JsonData   string     `json:"jsonData" gorm:"type:jsonb;not null"`
	IsShared   bool       `json:"isShared" gorm:"default:false;index"`
	SharedAt   *time.Time `json:"sharedAt" gorm:"index"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type UserPreferences struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	UserID    uint      `json:"userId" gorm:"uniqueIndex"`
	JsonData  string    `json:"jsonData" gorm:"type:jsonb;not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

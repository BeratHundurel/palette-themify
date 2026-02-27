package main

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:255;not null"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	GoogleID     string    `json:"googleId" gorm:"size:255;index"`
	AvatarURL    string    `json:"avatarUrl" gorm:"size:512"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Palettes     []Palette `json:"palettes" gorm:"foreignKey:UserID"`
}

type Palette struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"userId" gorm:"index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	JsonData  string    `json:"jsonData" gorm:"type:jsonb;not null"`
	IsSystem  bool      `json:"isSystem" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserPreferences struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"uniqueIndex"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	JsonData  string    `json:"jsonData" gorm:"type:jsonb;not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

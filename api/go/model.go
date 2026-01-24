package main

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:255;not null"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
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

type Workspace struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     *uint     `json:"userId" gorm:"index"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name       string    `json:"name" gorm:"size:255;not null"`
	JsonData   string    `json:"jsonData" gorm:"type:jsonb;not null"`
	ImageData  string    `json:"imageData" gorm:"type:text;not null"`
	ShareToken *string   `json:"shareToken" gorm:"size:64;uniqueIndex"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

package entity

import "time"

type User struct {
	Guid         string        `bson:"guid" json:"guid"`
	RefreshToken string        `bson:"refresh_token" json:"refresh_token"`
	IsValid      bool          `bson:"is_valid" json:"is_valid"`
	ExpAt        time.Duration `bson:"exp_at" json:"exp_at"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

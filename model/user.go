package model

type User struct {
	ID            int
	UserName      string `gorm:"column:username"`
	FirstName     string
	LastName      string
	ImageURL      string
	WithingsToken string
}

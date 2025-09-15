package userdata

type User struct {
	Phone        string `gorm:"column:phone_number;primaryKey"`
	PasswordHash string `gorm:"column:password_hash;type:varchar;not null"`
}

type UserRepo interface {
	Authorize(login, password string) (*User, error)
	Register(login, password string)
}

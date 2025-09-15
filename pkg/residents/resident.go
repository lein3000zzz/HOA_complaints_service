package residents

type Resident struct {
	ID       string  `gorm:"column:id;type:char(40);primaryKey"`
	Phone    string  `gorm:"column:phone_number;not null"`
	FullName string  `gorm:"column:full_name;not null"`
	Houses   []House `gorm:"many2many:resident_houses;"`
}

type House struct {
	ID        int        `gorm:"type:bigint;primaryKey"`
	Address   string     `gorm:"column:address;not null"`
	Residents []Resident `gorm:"many2many:resident_houses;"`
}

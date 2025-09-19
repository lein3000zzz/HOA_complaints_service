package residence

type Resident struct {
	ID       string  `gorm:"column:id;type:char(40);primaryKey"`
	Phone    string  `gorm:"type:varchar(40);column:phone_number;not null"`
	FullName string  `gorm:"type:varchar(40);column:full_name;not null"`
	Houses   []House `gorm:"many2many:residents_houses;"`
}

type House struct {
	ID        int        `gorm:"type:bigint;primaryKey"`
	Address   string     `gorm:"type:varchar(100);column:address;not null"`
	Residents []Resident `gorm:"many2many:residents_houses;"`
}

type ResidentHouse struct {
	ResidentID string `gorm:"column:id_resident;type:char(40);primaryKey"`
	HouseID    int    `gorm:"column:id_house;type:bigint;primaryKey"`
}

type ResidentsController interface {
	RegisterNewResident(phone, fullName string) (*Resident, error)
	RegisterNewHouse(address string) (*House, error)
	AddResidentAddressAssoc(residentID string, houseID int) error
	GetResidentByPhoneNumber(phoneNumber string) (*Resident, error)
	ValidateResidentHouse(residentID string, houseID int) (bool, error)
}

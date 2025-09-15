package residents

type ResidentPg Resident

func (ResidentPg) TableName() string {
	return "residents"
}

type HousePg House

func (HousePg) TableName() string {
	return "houses"
}

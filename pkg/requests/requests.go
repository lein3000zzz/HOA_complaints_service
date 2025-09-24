package requests

import "time"

type Request struct {
	ID             string        `gorm:"type:char(40);primaryKey"`
	ResidentID     string        `gorm:"column:id_resident;type:char(40);not null"`
	HouseID        int           `gorm:"column:id_house;type:bigint;not null"`
	RequestType    RequestType   `gorm:"column:type;type:request_type;not null"`
	Complaint      string        `gorm:"type:text;not null"`
	Cost           *float64      `gorm:"type:numeric(10,2)"`
	Status         RequestStatus `gorm:"type:request_status;default:'создана'"`
	ResponsibleID  *int          `gorm:"column:id_responsible;type:bigint"`
	OrganizationID *string       `gorm:"column:id_organization;type:char(40)"`
	CreatedAt      time.Time     `gorm:"column:created_at;type:timestamp;not null;default:now()"`
}

type RequestFilter struct {
	ID             *string
	ResidentID     *string
	HouseID        *int
	RequestType    *RequestType
	Complaint      *string
	Cost           *float64
	Status         *RequestStatus
	ResponsibleID  *int
	OrganizationID *string
	CreatedAt      *time.Time

	Limit  int
	Offset int
	Sort   string
}

type InitialRequestData struct {
	ResidentID  string
	HouseID     int
	RequestType RequestType
	Complaint   string
}

type RequestRepo interface {
	CreateRequest(requestData InitialRequestData) (*Request, error)
	GetResidentRequestsByPhone(phoneNumber string, limit, offset int, sort string) ([]*Request, int, error)
	GetAll(limit, offset int, sort string) ([]*Request, int, error)
	GetByFilter(filter RequestFilter) ([]*Request, int, error)
	//UpdateRequest(request Request) error
	//SelectResponsible(request Request)
}

type RequestType string

const (
	TypeApartmentInternal RequestType = "ремонт_внутриквартирный"
	TypeHouseCommon       RequestType = "ремонт_общедомового_имущества"
)

func (s RequestType) IsValid() bool {
	switch s {
	case TypeApartmentInternal, TypeHouseCommon:
		return true
	default:
		return false
	}
}

type RequestStatus string

const (
	StatusCreated     RequestStatus = "создана"
	StatusAssigned    RequestStatus = "назначена_исполнителю"
	StatusCompleted   RequestStatus = "выполнена"
	StatusCancelled   RequestStatus = "отменена"
	StatusSuspended   RequestStatus = "приостановлена"
	StatusTransferred RequestStatus = "передана_организации"
)

func (s RequestStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusAssigned, StatusCompleted, StatusCancelled, StatusSuspended, StatusTransferred:
		return true
	default:
		return false
	}
}

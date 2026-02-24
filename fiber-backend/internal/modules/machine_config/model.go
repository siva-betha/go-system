package machine_config

import (
	"time"
)

type Machine struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" validate:"required"`
	IP        string    `json:"ip" validate:"required,ip"`
	AmsNetID  string    `json:"ams_net_id" validate:"required"`
	Port      int       `json:"port" validate:"required"`
	Chambers  []Chamber `json:"chambers" gorm:"foreignKey:MachineID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Chamber struct {
	ID        string   `json:"id" gorm:"primaryKey"`
	MachineID string   `json:"machine_id"`
	Name      string   `json:"name" validate:"required"`
	Symbols   []Symbol `json:"symbols" gorm:"foreignKey:ChamberID;constraint:OnDelete:CASCADE"`
}

type Symbol struct {
	ID        string `json:"id" gorm:"primaryKey"`
	ChamberID string `json:"chamber_id"`
	Name      string `json:"name" validate:"required"`
	DataType  string `json:"data_type" validate:"required"`
	Unit      string `json:"unit,omitempty"`
}

type MachineResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	IP       string            `json:"ip"`
	AmsNetID string            `json:"ams_net_id"`
	Port     int               `json:"port"`
	Status   string            `json:"status"` // Configured/Online/Offline
	Chambers []ChamberResponse `json:"chambers"`
}

type ChamberResponse struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Symbols []SymbolResponse `json:"symbols"`
}

type SymbolResponse struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Unit     string `json:"unit,omitempty"`
}

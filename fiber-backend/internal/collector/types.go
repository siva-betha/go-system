package collector

import "time"

type MachineConfig struct {
    ID        string          `json:"id" gorm:"primaryKey"`
    Name      string          `json:"name"`
    IP        string          `json:"ip"`
    AmsNetID  string          `json:"ams_net_id"`
    Port      int             `json:"port"`
    Chambers  []ChamberConfig `json:"chambers" gorm:"foreignKey:MachineID"`
    CreatedAt time.Time       `json:"created_at"`
    UpdatedAt time.Time       `json:"updated_at"`
}

type ChamberConfig struct {
    ID        string         `json:"id" gorm:"primaryKey"`
    MachineID string         `json:"machine_id"`
    Name      string         `json:"name"`
    Symbols   []SymbolConfig `json:"symbols" gorm:"foreignKey:ChamberID"`
}

type SymbolConfig struct {
    ID        string `json:"id" gorm:"primaryKey"`
    ChamberID string `json:"chamber_id"`
    Name      string `json:"name"` // PLC variable name (e.g., "GVL.temperature")
    DataType  string `json:"data_type"`
    Unit      string `json:"unit,omitempty"`
}

type PLCData struct {
    MachineID   string      `json:"machine_id"`
    ChamberID   string      `json:"chamber_id"`
    Symbol      string      `json:"symbol"`
    Value       interface{} `json:"value"`
    Quality     int         `json:"quality"`
    Timestamp   time.Time   `json:"timestamp"`
    SequenceNum uint64      `json:"sequence_num"`
}

type StatsRequest struct {
    MachineID string `json:"machine_id"`
}

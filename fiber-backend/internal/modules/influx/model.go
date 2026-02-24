package influx

type Point struct {
	Time               string      `json:"time"`
	Value              interface{} `json:"value"`
	Field              string      `json:"field"`
	Measurement        string      `json:"measurement"`
	ChamberID          string      `json:"chamber_id"`
	Destination        string      `json:"destination"`
	LayerID            string      `json:"layer_id"`
	Source             string      `json:"source"`
	System             string      `json:"system"`
	TelegrafInstanceID string      `json:"telegraf_instance_id"`
	WaferID            string      `json:"wafer_id"`
}

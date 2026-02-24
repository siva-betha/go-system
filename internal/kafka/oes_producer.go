package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type OESProducer struct {
	producer *kafka.Producer
	topic    string
}

type Spectrum struct {
	Timestamp      time.Time
	SequenceNum    uint32
	AcquisitionCtr uint16
	SegmentCtr     uint16
	Intensities    []uint16
	Metadata       map[string]interface{}
}

type RecipeContext struct {
	RecipeID    string
	ProcessJob  string
	SubstrateID string
	StepIndex   int
}

func NewOESProducer(broker string, topic string) (*OESProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return nil, err
	}
	return &OESProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *OESProducer) PublishSpectrum(spectrum *Spectrum, recipe *RecipeContext) error {
	value := map[string]interface{}{
		"timestamp":       spectrum.Timestamp.UnixNano(),
		"sequence_num":    spectrum.SequenceNum,
		"acquisition_ctr": spectrum.AcquisitionCtr,
		"segment_ctr":     spectrum.SegmentCtr,
		"intensities":     spectrum.Intensities,
		"metadata":        spectrum.Metadata,
	}

	if recipe != nil {
		value["recipe_id"] = recipe.RecipeID
		value["process_job"] = recipe.ProcessJob
		value["substrate_id"] = recipe.SubstrateID
		value["recipe_step"] = recipe.StepIndex
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	key := "default"
	if recipe != nil {
		key = fmt.Sprintf("%s_%s", recipe.ProcessJob, recipe.SubstrateID)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: payload,
	}

	return p.producer.Produce(msg, nil)
}

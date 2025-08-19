package main

import (
	"authz/internal/engine/rbac"
	"authz/internal/entity"
	"authz/internal/pkg"
	"authz/internal/service"
	"authz/internal/service/database"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var graph *rbac.Graph
var lastKey int64

func main() {
	if err := pkg.NewConfig(); err != nil {
		panic(err)
	}

	lc := pkg.NewLifecycleParallel()
	r := service.NewKafkaReader(lc)
	c := context.Background()
	graph = rbac.NewGraph()
	if err := r.SetOffset(kafka.FirstOffset); err != nil {
		panic(err)
	}
	for {
		readCtx, cancel := context.WithTimeout(c, 3*time.Second)
		defer cancel()
		m, err := r.ReadMessage(readCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				break
			}
			log.Error().Err(err).Msg("Failed to read message")
		}
		if err := applyMessage(m); err != nil {
			log.Error().Err(err).Msg("Failed to apply message")
		}
	}
	fmt.Println("lastKey: ", lastKey)
	// fmt.Println(len(graph))
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	start := time.Now()
	if err := enc.Encode(&graph); err != nil {
		panic(err)
	}
	marshalTime := time.Since(start)
	marshaledData := buf.Bytes()
	size := len(marshaledData)
	fmt.Printf("Marshal Time: %s\n", marshalTime)
	fmt.Printf("Size: %d bytes\n", size)

	db, err := database.New(lc)
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&rbac.GraphCheckpoint{})
	start = time.Now()
	if err := db.Create(&rbac.GraphCheckpoint{
		LastOffset: lastKey,
		Data:       marshaledData,
	}).Error; err != nil {
		panic(err)
	}

	saveTime := time.Since(start)
	fmt.Printf("Save Time: %s\n\n", saveTime)

	loadPoint := rbac.GraphCheckpoint{}
	start = time.Now()
	if err := db.First(&loadPoint).Error; err != nil {
		panic(err)
	}
	loadTime := time.Since(start)
	fmt.Printf("Load Time: %s\n\n", loadTime)

	graph = rbac.NewGraph()
	dec := gob.NewDecoder(bytes.NewReader(marshaledData))
	start = time.Now()
	if err := dec.Decode(&graph); err != nil {
		panic(err)
	}
	unmarshalTime := time.Since(start)
	fmt.Printf("Unmarshal Time: %s\n\n", unmarshalTime)

}

func applyMessage(m kafka.Message) error {
	type Val struct {
		rbac.Tuple
		Op string `json:"__op"`
	}
	var val Val

	if err := json.Unmarshal(m.Value, &val); err != nil {
		return err
	}

	lastKey = m.Offset

	sbj := entity.Instance{Ns: val.SbjNs, Id: val.SbjId}
	rel := val.Relation
	obj := entity.Instance{Ns: val.ObjNs, Id: val.ObjId}

	// Apply operation type from val.Op ("c"=create, "d"=delete, etc.)
	switch val.Op {
	case "c": // create/add edge
		graph.Create(obj, rel, sbj)
	case "d": // delete/remove edge
		graph.Delete(obj, rel, sbj)
	default:
		// handle other ops if any
	}

	return nil
}

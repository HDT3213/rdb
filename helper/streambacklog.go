package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

// StreamBacklogReport writes a CSV report of the pending backlog of every consumer group of every stream in the rdb file
func StreamBacklogReport(rdbFilename string, output *os.File, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}
	_, err = output.WriteString("database,key,group,consumer,pending,oldest_delivery_time,max_delivery_count,undelivered,dangling\n")
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()
	var writeErr error
	err = dec.Parse(func(object model.RedisObject) bool {
		stream, ok := object.(*model.StreamObject)
		if !ok {
			return true
		}
		for _, row := range stream.ConsumerBacklog() {
			writeErr = csvWriter.Write([]string{
				strconv.Itoa(object.GetDBIndex()),
				object.GetKey(),
				row.Group,
				row.Consumer,
				strconv.Itoa(row.Pending),
				strconv.FormatUint(row.OldestDeliveryTime, 10),
				strconv.FormatUint(row.MaxDeliveryCount, 10),
				strconv.Itoa(row.Undelivered),
				strconv.Itoa(row.Dangling),
			})
			if writeErr != nil {
				return false
			}
		}
		return true
	})
	if writeErr != nil {
		return fmt.Errorf("csv write failed: %v", writeErr)
	}
	if err != nil {
		return err
	}
	csvWriter.Flush()
	if err = csvWriter.Error(); err != nil {
		return fmt.Errorf("csv write failed: %v", err)
	}
	return nil
}

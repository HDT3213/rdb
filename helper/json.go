package helper

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

var jsonEncoder = sonic.ConfigDefault

// ConcurrentOption sets the number of goroutines for json converter
type ConcurrentOption int

// WithConcurrent sets the number of goroutines for json converter
func WithConcurrent(c int) ConcurrentOption {
	return ConcurrentOption(c)
}

// ToJsons read rdb file and convert to json file
func ToJsons(rdbFilename string, jsonFilename string, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	if jsonFilename == "" {
		return errors.New("output file path is required")
	}
	// open file
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	jsonFile, err := os.Create(jsonFilename)
	if err != nil {
		return fmt.Errorf("create json %s failed, %v", jsonFilename, err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()
	// create decoder
	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}
	// parse rdb
	_, err = jsonFile.WriteString("[\n")
	if err != nil {
		return fmt.Errorf("write json  failed, %v", err)
	}

	// parse options
	concurrent := 1
	cpuNum := runtime.NumCPU()
	if cpuNum > 1 {
		concurrent = cpuNum - 1 // leave one core for parser
	}
	for _, opt := range options {
		switch o := opt.(type) {
		case ConcurrentOption:
			concurrent = int(o)
		}
	}

	redisObjectBuffer := make(chan model.RedisObject, 1000)
	jsonStringBuffer := make(chan []byte, 1000)

	// parser goroutine
	empty := true
	go func() {
		err = dec.Parse(func(object model.RedisObject) bool {
			redisObjectBuffer <- object
			return true
		})
		close(redisObjectBuffer)
	}()
	// json marshaller goroutine
	wg := &sync.WaitGroup{}
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			for object := range redisObjectBuffer {
				data, err := jsonEncoder.Marshal(object) // enable SortMapKeys to ensure same result
				if err != nil {
					fmt.Printf("json marshal failed: %v", err)
					continue
				}
				jsonStringBuffer <- data
			}
			wg.Done()
		}()
	}
	// write goroutine
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		for data := range jsonStringBuffer {
			data = append(data, ',', '\n')
			_, err = jsonFile.Write(data)
			if err != nil {
				fmt.Printf("write failed: %v", err)
				continue
			}
			empty = false
		}
		wg2.Done()
	}()

	wg.Wait()
	close(jsonStringBuffer)
	wg2.Wait() // wait writing goroutine

	// finish json
	if !empty {
		_, err = jsonFile.Seek(-2, 2)
		if err != nil {
			return fmt.Errorf("error during seek in file: %v", err)
		}
	}
	_, err = jsonFile.WriteString("\n]")
	if err != nil {
		return fmt.Errorf("error during write in file: %v", err)
	}
	return nil
}

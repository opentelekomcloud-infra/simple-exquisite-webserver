package main

import (
	"github.com/twinj/uuid"
	"log"
	"math/rand"
	"time"
	"unsafe"
)

type Entity struct {
	Uuid string `json:"uuid"`
	Data string `json:"data"`
}

const DataRandCS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ :;~`\\|/?.,<>{}()&*%$#@"

var src = rand.NewSource(time.Now().UnixNano())

func randomByteSlice(size int, prefix string, charset string) []byte {
	csLen := len(charset)
	prefLen := len(prefix)
	result := make([]byte, size)
	copy(result, prefix)
	for i := prefLen; i < size; i++ {
		result[i] = charset[src.Int63()%int64(csLen)]
	}
	return result
}

func RandomString(size int, prefix string, charset ...string) string {
	cs := DataRandCS
	if len(charset) > 0 {
		cs = charset[0]
	}
	result := randomByteSlice(size, prefix, cs)
	return *(*string)(unsafe.Pointer(&result)) // faster way to convert big slice to string
}

//GenerateSomeEntities create `count` of random entities with given chars in data (20000 by default)
func GenerateSomeEntities(count int, dataSize ...int) []Entity {
	size := 20000
	if len(dataSize) > 0 {
		size = dataSize[0]
	}

	var data = make([]Entity, count)
	startTime := time.Now().Unix()
	for i := 0; i < count; i++ {
		func() {
			data[i] = Entity{
				Uuid: uuid.NewV4().String(),
				Data: RandomString(size, "RANDOM DATA: "),
			}
		}()
	}
	log.Print("Waiting for data to be generated...")
	log.Printf("Generated data in %vs", time.Now().Unix()-startTime)
	return data
}

package database

import (
	"log"
	"testing"
)

func TestEncodeKey(t *testing.T) {
	result1, _ := EncodeKey(1, []byte{2, 3, 4, 5}, []byte{234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111, 234, 121, 123, 111})
	log.Println(len(result1))

	result2, _ := EncodeKey(2, []byte{234, 121, 123, 111})
	log.Println(len(result2))

	result3, _ := EncodeKey(3, []byte{234}, []byte{5})
	log.Println(len(result3))
}

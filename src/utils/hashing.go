package utils

import (
	"encoding/binary"
	"hash/fnv"
	"strings"

	"github.com/schollz/mnemonicode"
)

func StringToReadableHash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, h.Sum32())
	result := mnemonicode.EncodeWordList([]string{}, bs)
	return strings.Join(result, "-")
}

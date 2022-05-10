package crypt

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"math/rand"
)

func MD5(str string) string {
	c := md5.New()
	_, err := c.Write([]byte(str))
	if nil != err {
		return ""
	}
	return hex.EncodeToString(c.Sum(nil))
}

func SHA1(str string) string {
	c := sha1.New()
	_, err := c.Write([]byte(str))
	if nil != err {
		return ""
	}
	return hex.EncodeToString(c.Sum(nil))
}

func CRC32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

// RangeRandFloat generate a random float number
func RangeRandFloat() float64 {
	return rand.Float64()
}

func RangeRandInt() int {
	return rand.Intn(2)
}

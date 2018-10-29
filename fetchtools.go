package qqwry

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net/http"
)

const (
	COPY_WRITE_URL = "http://update.cz88.net/ip/copywrite.rar"
	QQWRY_URL      = "http://update.cz88.net/ip/qqwry.rar"
)

func GetContent(url string) (b []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	return
}

func GetKey(b []byte) (key uint32, err error) {
	if len(b) != 280 {
		return 0, errors.New("copywrite.rar is corrupt")
	}
	key = binary.LittleEndian.Uint32(b[20:])
	return
}

func Decrypt(b []byte, key uint32) (_ []byte, err error) {
	for i := 0; i < 0x200; i++ {
		key *= uint32(0x805)
		key++
		key &= uint32(0xff)
		b[i] = b[i] ^ byte(key)
	}
	rc, err := zlib.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}

func GetOnlineQQwryDat() (b []byte, err error) {
	var copyWriteData, qqwryData []byte
	if copyWriteData, err = GetContent(COPY_WRITE_URL); err != nil {
		return
	}
	if qqwryData, err = GetContent(QQWRY_URL); err != nil {
		return
	}
	var key uint32
	if key, err = GetKey(copyWriteData); err != nil {
		return
	}
	if b, err = Decrypt(qqwryData, key); err != nil {
		return
	}
	return
}

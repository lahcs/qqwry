package qqwry

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	COPY_WRITE_URL = "http://update.cz88.net/ip/copywrite.rar"
	QQWRY_URL      = "http://update.cz88.net/ip/qqwry.rar"

	COPY_WRITE_FILE = "copywrite.dat"
	QQWRY_FILE      = "qqwry.dat"
)

const (
	HTTP_CONNECT_TIMEOUT time.Duration = time.Second * 5
	HTTP_READ_TIMEOUT    time.Duration = time.Second * 60
)

func NewClient() *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(HTTP_READ_TIMEOUT)
				c, err := net.DialTimeout(netw, addr, HTTP_CONNECT_TIMEOUT)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			DisableKeepAlives: true,
		},
	}
}

func GetContent(url string) (b []byte, err error) {
	client := NewClient()

	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(reqest)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	return
}

func GetKey(b []byte) (key uint32, err error) {
	if len(b) != 280 {
		return 0, errors.New("Copywrite.rar is Corrupt")
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
	log.Printf("[QQWry]Load Online Database...\n")

	var copyWriteData, qqwryData []byte
	if copyWriteData, err = GetContent(COPY_WRITE_URL); err != nil {
		return
	} else {
		ioutil.WriteFile(COPY_WRITE_FILE, copyWriteData, 0777)
	}

	if qqwryData, err = GetContent(QQWRY_URL); err != nil {
		return
	} else {
		ioutil.WriteFile(QQWRY_FILE, qqwryData, 0777)
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

func GetLocalQQwryDat() (b []byte, err error) {
	log.Printf("[QQWry]Load Local Database...\n")

	var copyWriteData, qqwryData []byte
	if copyWriteData, err = ioutil.ReadFile(COPY_WRITE_FILE); err != nil {
		return
	}
	if qqwryData, err = ioutil.ReadFile(QQWRY_FILE); err != nil {
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

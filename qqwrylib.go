package qqwry

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/yinheli/mahonia"
)

const (
	INDEX_LEN       = 7
	REDIRECT_MODE_1 = 0x01
	REDIRECT_MODE_2 = 0x02
)

type QQwry struct {
	buff  []byte
	start uint32
	end   uint32
	sync.RWMutex
}

type Rq struct {
	Ip   string
	City string
	Area string
}

var GlobalQQwry *QQwry
var Once = &sync.Once{}

func Find(ip string) (result *Rq, err error) {
	Once.Do(func() {
		if GlobalQQwry, err = NewQQwry(); err != nil {
			log.Printf("init default qqwry error, err=%s", err.Error())
		} else {
			log.Printf("init default qqwry success")
		}
	})
	result, err = GlobalQQwry.Find(ip)
	return
}

func NewQQwry() (qqwry *QQwry, err error) {
	qqwry = &QQwry{}
	if err = qqwry.setBuff(); err != nil {
		log.Printf("set buff error, err=%s\n", err.Error())
		return
	}
	qqwry.update()
	return
}

func (this *QQwry) update() {
	go func() {
		for true {
			// 一小时探测一次, 太频繁会被屏蔽掉
			time.Sleep(1 * time.Hour)
			this.setBuff()
		}
	}()
}

func (this *QQwry) setBuff() (err error) {
	defer func() {
		if err != nil {
			log.Printf("set buff error, err=%s\n", err.Error())
		} else {
			log.Printf("set buff success")
		}
	}()
	var buff []byte
	if buff, err = GetOnlineQQwryDat(); err != nil {
		return
	}
	if bytes.Compare(buff, this.buff) == 0 {
		log.Printf("don't update")
		return
	}
	if len(buff) < 9 {
		err = errors.New("invalid buff")
		return
	}
	start := binary.LittleEndian.Uint32(buff[:4])
	end := binary.LittleEndian.Uint32(buff[4:8])
	this.Lock()
	defer this.Unlock()
	this.buff = buff
	this.start = start
	this.end = end
	return
}

// func NewQQwry(file string) (qqwry *QQwry) {
// 	qqwry = &QQwry{}
// 	f, e := os.Open(file)
// 	if e != nil {
// 		log.Println(e)
// 		return nil
// 	}
// 	defer f.Close()
// 	qqwry.buff, e = ioutil.ReadAll(f)
// 	if e != nil {
// 		log.Println(e)
// 		return nil
// 	}
// 	qqwry.start = binary.LittleEndian.Uint32(qqwry.buff[:4])
// 	qqwry.end = binary.LittleEndian.Uint32(qqwry.buff[4:8])
// 	return qqwry
// }

func (this *Rq) String() string {

	d, _ := json.Marshal(this)
	return string(d)
}
func (this *QQwry) Find(ip string) (result *Rq, err error) {
	result = &Rq{Ip: ip}
	defer func() {
		if err := recover(); err != nil {
			result.City = "未知"
			result.Area = "未知"
		}
	}()

	var city []byte
	var area []byte
	ip_1 := net.ParseIP(ip)
	if ip_1 == nil {
		err = errors.New("invalid ip")
		return
	}

	this.RLock()
	defer this.RUnlock()
	offset := this.searchRecord(binary.BigEndian.Uint32(ip_1.To4()))
	if offset <= 0 {
		err = errors.New("not found")
		return
	}
	mode := this.readMode(offset + 4)
	if mode == REDIRECT_MODE_1 {
		cityOffset := this.readUint32FromByte3(offset + 5)

		mode = this.readMode(cityOffset)
		if mode == REDIRECT_MODE_2 {
			c := this.readUint32FromByte3(cityOffset + 1)
			city = this.readString(c)
			cityOffset += 4
			area = this.readArea(cityOffset)

		} else {
			city = this.readString(cityOffset)
			cityOffset += uint32(len(city) + 1)
			area = this.readArea(cityOffset)
		}

	} else if mode == REDIRECT_MODE_2 {
		cityOffset := this.readUint32FromByte3(offset + 5)
		city = this.readString(cityOffset)
		area = this.readArea(offset + 8)
	}
	enc := mahonia.NewDecoder("gbk")
	result.City = enc.ConvertString(string(city))
	result.Area = enc.ConvertString(string(area))
	return
}

func (this *QQwry) readUint32FromByte3(offset uint32) uint32 {
	return byte3ToUInt32(this.buff[offset : offset+3])
}
func (this *QQwry) readMode(offset uint32) byte {
	return this.buff[offset : offset+1][0]
}

func (this *QQwry) readString(offset uint32) []byte {

	i := 0
	for {

		if this.buff[int(offset)+i] == 0 {
			break
		} else {
			i++
		}

	}
	return this.buff[offset : int(offset)+i]
}

func (this *QQwry) readArea(offset uint32) []byte {
	mode := this.readMode(offset)
	if mode == REDIRECT_MODE_1 || mode == REDIRECT_MODE_2 {
		areaOffset := this.readUint32FromByte3(offset + 1)
		if areaOffset == 0 {
			return []byte("")
		} else {
			return this.readString(areaOffset)
		}
	} else {
		return this.readString(offset)
	}
	return []byte("")
}

func (this *QQwry) getRecord(offset uint32) []byte {
	return this.buff[offset : offset+INDEX_LEN]
}

func (this *QQwry) getIPFromRecord(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf[:4])
}

func (this *QQwry) getAddrFromRecord(buf []byte) uint32 {
	return byte3ToUInt32(buf[4:7])
}

func (this *QQwry) searchRecord(ip uint32) uint32 {

	start := this.start
	end := this.end

	// log.Printf("len info %v, %v ---- %v, %v", start, end, hex.EncodeToString(header[:4]), hex.EncodeToString(header[4:]))
	for {
		mid := this.getMiddleOffset(start, end)
		buf := this.getRecord(mid)
		_ip := this.getIPFromRecord(buf)

		// log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))

		if end-start == INDEX_LEN {
			//log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))
			offset := this.getAddrFromRecord(buf)
			buf = this.getRecord(mid + INDEX_LEN)
			if ip < this.getIPFromRecord(buf) {
				return offset
			} else {
				return 0
			}
		}

		// 找到的比较大，向前移
		if _ip > ip {
			end = mid
		} else if _ip < ip { // 找到的比较小，向后移
			start = mid
		} else if _ip == ip {
			return byte3ToUInt32(buf[4:7])
		}

	}
	return 0
}

func (this *QQwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / INDEX_LEN) >> 1
	return start + records*INDEX_LEN
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

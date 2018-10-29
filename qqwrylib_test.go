package qqwry

import (
	"testing"
)

func TestGlobalQuery(t *testing.T) {
	var result *Rq
	var err error
	if result, err = Find("202.106.0.20"); err != nil {
		t.Error(err.Error())
	}
	if result.City != "北京市" {
		t.Error(result.City + " != 北京市")
	}
}

func TestQuery(t *testing.T) {
	var q *QQwry
	var err error
	if q, err = NewQQwry(); err != nil {
		t.Error(err.Error())
	}
	var result *Rq
	if result, err = q.Find("202.106.0.20"); err != nil {
		t.Error(err.Error())
	}
	if result.City != "北京市" {
		t.Error(result.City + " != 北京市")
	}
}

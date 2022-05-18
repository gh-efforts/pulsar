package model

import "time"

type AllType int8

const (
	DefaultAllType AllType = 1 //all addresses of  all methods
)

type UserAppSubAll struct {
	AppId      string `bson:"app_id"`
	AllType    int8   `bson:"all_type"`
	UpdateTime int64  `bson:"update_time"`
	CreateTime int64  `bson:"create_time"`
}

func NewDefaultAppSubAll() UserAppSubAll {
	now := time.Now().Unix()
	return UserAppSubAll{
		UpdateTime: now,
		CreateTime: now,
		AllType:    int8(DefaultAllType),
	}
}

func (watch *UserAppSubAll) IsEmpty() bool {
	return watch.AppId == ""
}

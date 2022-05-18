package model

import "time"

type UserAppSub struct {
	AppId      string `bson:"app_id"`
	Address    string `bson:"address"`
	UpdateTime int64  `bson:"update_time"`
	CreateTime int64  `bson:"create_time"`
	State      int8   `bson:"state"`
}

type SpecialUserAppSub struct {
	AppId   string `bson:"app_id"`
	Address string `bson:"address"`
}

func NewDefaultAppSub() UserAppSub {
	now := time.Now().Unix()
	return UserAppSub{
		UpdateTime: now,
		CreateTime: now,
		State:      1,
	}
}

func (watch *UserAppSub) IsEmpty() bool {
	return watch.AppId == ""
}

package model

import "time"

type UserAppWatch struct {
	AppId      string `bson:"app_id"`
	Address    string `bson:"address"`
	UpdateTime int64  `bson:"update_time"`
	CreateTime int64  `bson:"create_time"`
	State      int8   `bson:"state"`
}

type SpecialUserAppWatch struct {
	AppId   string `bson:"app_id"`
	Address string `bson:"address"`
}

func NewDefaultAppWatch() UserAppWatch {
	now := time.Now().Unix()
	return UserAppWatch{
		UpdateTime: now,
		CreateTime: now,
		State:      1,
	}
}

func (watch *UserAppWatch) IsEmpty() bool {
	return watch.AppId == ""
}

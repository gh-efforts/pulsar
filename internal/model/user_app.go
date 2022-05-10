package model

import (
	"time"

	"github.com/bitrainforest/pulsar/internal/utils/crypt"

	uuid "github.com/satori/go.uuid"
)

type (
	AppType  int8
	AppState int8
	// todo we may have user first, then app??

	UserApp struct {
		// todo add userId
		AppId      string `bson:"app_id"`
		AppSecret  string `bson:"app_secret"`
		AppType    int8   `bson:"app_type"`
		State      int8   `bson:"state"`
		UpdateTime int64  `bson:"update_time"`
		CreateTime int64  `bson:"create_time"`
	}
)

const (
	SubAppType AppType = 1 //for sub message

	DefaultState AppState = 1
)

func NewDefaultApp() UserApp {
	now := time.Now().Unix()
	app := UserApp{
		AppId:      "",
		CreateTime: now,
		UpdateTime: now,
		AppType:    int8(SubAppType),
		State:      int8(DefaultState),
	}
	app.GenAppId()
	app.GenAppSecret()
	return app
}

func (app *UserApp) IsEmpty() bool {
	return app.AppId == ""
}

func (app *UserApp) GenAppId() {
	app.AppId = uuid.NewV4().String()
}

func (app *UserApp) GenAppSecret() {
	app.AppSecret = crypt.SHA1(app.AppId)
}

func (app *UserApp) CheckAppSecret(appSecret string) bool {
	return app.AppSecret == appSecret
}

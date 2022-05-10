package req

type AppReq struct {
	AppId     string `json:"appId" binding:"required,gt=0"`
	AppSecret string `json:"appSecret" binding:"required,gt=0"`
}

type ApplyResp struct {
	AppId     string `json:"appId"`
	AppSecret string `json:"appSecret"`
}

type AddSubReq struct {
	Address string `json:"address" binding:"required,gt=0"`
}

type AddSubResp struct {
}

type CancelSubAddressReq struct {
	Address string `json:"address" binding:"required,gt=0"`
}

type CancelSubAddressResp struct {
}

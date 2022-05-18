package req

type AppReq struct {
	AppId     string `json:"appId" binding:"required,gt=0"`
	AppSecret string `json:"appSecret" binding:"required,gt=0,lt=200"`
}

type ApplyResp struct {
	AppId     string `json:"appId"`
	AppSecret string `json:"appSecret"`
}

type AddSubReq struct {
	SubType int8   `json:"subType" binding:"required,gte=1,lte=2"` //1 all address
	Address string `json:"address" binding:"required_if=SubType 2,omitempty,gt=0,lte=300"`
}

func (addSub *AddSubReq) IsAll() bool {
	return addSub.SubType == 1
}

type AddSubResp struct {
}

type CancelSubAddressReq struct {
	SubType int8   `json:"subType" binding:"required,gte=1,lte=2"`
	Address string `json:"address" binding:"required_if=SubType 2,omitempty,gt=0,lte=300"`
}

func (receiver CancelSubAddressReq) IsAll() bool {
	return receiver.SubType == 1
}

type CancelSubAddressResp struct {
}

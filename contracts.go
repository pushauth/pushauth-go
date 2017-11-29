package pushauth

type ReqData struct {
	PublicKey string `json:"pk"`
	Data      string `json:"data"`
}

type Req struct {
	Mode          AuthMode `json:"mode"`
	Code          string   `json:"code"`
	FlashResponse bool     `json:"flash_response"`
}

type ReqSingle struct {
	Req
	AddrTo string `json:"addr_to"`
}

type ReqMultiple struct {
	Req
	AddrTo map[string]string `json:"addr_to"`
}

type ReqResp struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ReqDataResp struct {
	ReqHash string `json:"req_hash"`
	Answer  bool   `json:"answer"`
}

type QRImage struct {
	Size            int    `json:"size"`
	Color           string `json:"color"`
	BackgroundColor string `json:"backgroundColor"`
	Margin          int    `json:"margin"`
}

type QRRequest struct {
	Image QRImage `json:"image"`
}

type QRResponse struct {
	ReqHash string `json:"req_hash"`
	QRurl   string `json:"qr_url"`
}

type CheckRequest struct {
	ReqHash string `json:"req_hash"`
}

type StatusResp struct {
	Answered        bool   `json:"answered"`
	Answer          bool   `json:"answer"`
	ResponseCode    int    `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	ResponseDT      int    `json:"response_dt"`
}

type StatusRespWait struct {
	*StatusResp
	Error error
}

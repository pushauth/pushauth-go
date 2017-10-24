package pushauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	PushURL   = "https://api.pushauth.io/push/send"
	QRShowURL = "https://api.pushauth.io/qr/show"
	StatusURL = "https://api.pushauth.io/push/status"

	AppJSONMime = "application/json"
)

var (
	// HMAC Error
	ErrorHMACInvalid = errors.New("invalid hmac")

	// Status Waiter Error
	ErrorStatusWaitClosed = errors.New("wait ended")

	// HTTP API Errors
	ErrorAccessDenied        = errors.New("access denied")
	ErrorNotFound            = errors.New("not found")
	ErrorMethodNotAllowed    = errors.New("method not allowed")
	ErrorUnprocessableEntity = errors.New("unprocessable entity")
	ErrorInternalServerError = errors.New("internal server error")
)

type PushAuth struct {
	WaitTime              time.Duration
	publicKey, privateKey []byte
	hash                  hash.Hash
}

func encodeBase64(msg []byte) string {
	return base64.StdEncoding.EncodeToString(msg)
}

func decodeBase64(msg []byte) []byte {
	byts, err := base64.StdEncoding.DecodeString(string(msg))
	if err != nil {
		return nil
	}
	return byts
}

func checkStatus(code int) error {
	switch code {
	case http.StatusForbidden:
		return ErrorAccessDenied
	case http.StatusNotFound:
		return ErrorNotFound
	case http.StatusMethodNotAllowed:
		return ErrorMethodNotAllowed
	case http.StatusUnprocessableEntity:
		return ErrorUnprocessableEntity
	case http.StatusInternalServerError:
		return ErrorInternalServerError
	default:
		return nil
	}
}

func NewPushAuth(publicKey, privateKey []byte, waitTime time.Duration) *PushAuth {
	return &PushAuth{publicKey: publicKey, privateKey: privateKey, hash: hmac.New(sha256.New, privateKey), WaitTime: waitTime}
}

func GetWaiterChans() (chan *StatusRespWait, chan struct{}) {
    return make(chan *StatusRespWait, 1), make(chan struct{}, 1)
}

func (p *PushAuth) getHMAC(data []byte) []byte {
	p.hash.Reset()
	p.hash.Write(data)
	return p.hash.Sum(nil)
}

func (p *PushAuth) encodeData(data interface{}) []byte {
	bts, err := json.Marshal(data)

	if err != nil {
		return nil
	}
	msg := encodeBase64(bts)
	reqData := &ReqData{PublicKey: string(p.publicKey),
		Data: fmt.Sprintf("%s.%s", msg, encodeBase64(p.getHMAC([]byte(msg))))}

	marshalized, _ := json.Marshal(reqData)
	return marshalized
}

func (p *PushAuth) decodeData(data []byte, out interface{}) error {
	splits := bytes.Split(data, []byte("."))

	generatedHMAC := encodeBase64(p.getHMAC(splits[0]))

	if !hmac.Equal(splits[1], []byte(generatedHMAC)) {
		return ErrorHMACInvalid
	}

	decoded := decodeBase64(splits[0])

	err := json.Unmarshal(decoded, out)

	if err != nil {
		return err
	}

	return nil
}

func (p *PushAuth) basicRequest(url, contentType string, data io.Reader, waitResponse bool) (*ReqDataResp, error) {
	respData, err := doPostRequest(url, contentType, data)
	if err != nil {
		return nil, err
	}

	respParsed := &ReqDataResp{}

	if err = p.decodeData([]byte(respData.Data), respParsed); err != nil {
		return nil, err
	}

	return respParsed, nil
}

func (p *PushAuth) PushSingle(to string, flashResponse bool) (*ReqDataResp, error) {
	req := Req{Mode: "push", FlashResponse: flashResponse}
	dPlain := ReqSingle{AddrTo: to, Req: req}
	dPush := p.encodeData(dPlain)

	reader := bytes.NewReader(dPush)

	return p.basicRequest(PushURL, AppJSONMime, reader, !flashResponse)
}

func (p *PushAuth) PushMult(to []string, flashResponse bool) (*ReqDataResp, error) {
	req := Req{Mode: "push", FlashResponse: flashResponse}
	mapTo := make(map[string]string)
	for idx := range to {
		mapTo[strconv.Itoa(idx+1)] = to[idx]
	}
	dPlain := ReqMultiple{AddrTo: mapTo, Req: req}
	dPush := p.encodeData(dPlain)
	strPush := string(dPush)
	_ = strPush

	reader := bytes.NewReader(dPush)

	return p.basicRequest(PushURL, AppJSONMime, reader, !flashResponse)
}

func (p *PushAuth) CodeSingle(to, code string) (*ReqDataResp, error) {
	req := Req{Mode: "code", Code: code}
	dPlain := ReqSingle{AddrTo: to, Req: req}
	dPush := p.encodeData(dPlain)

	reader := bytes.NewReader(dPush)

	resp, err := p.basicRequest(PushURL, AppJSONMime, reader, false)
	if err != nil {
		return nil, err
	}

	resp.Answer = true
	return resp, nil
}

func (p *PushAuth) WaitForStatus(hash string, out chan<- *StatusRespWait, closer <-chan struct{}) {
	data := p.encodeData(&CheckRequest{ReqHash: hash})

	ticker := time.NewTicker(p.WaitTime)
	for {
		select {
		case <-ticker.C:
			resp, err := doPostRequest(StatusURL, AppJSONMime, bytes.NewReader(data))
			if err != nil {
				out <- &StatusRespWait{&StatusResp{}, err}
				return
			}
			status := &StatusResp{}
			if err := p.decodeData([]byte(resp.Data), status); err != nil {
				out <- &StatusRespWait{&StatusResp{}, err}
				return
			}

			if !status.Answered || status.ResponseCode != 200 {
				continue
			}

			out <- &StatusRespWait{status, nil}
			return
		case <-closer:
			out <- &StatusRespWait{nil, ErrorStatusWaitClosed}
			return
		}
	}
}

func doPostRequest(url, contentType string, data io.Reader) (*ReqResp, error) {
	resp, err := http.Post(url, contentType, data)

	if err != nil {
		return nil, err
	}

	bts, _ := ioutil.ReadAll(resp.Body)

	err = checkStatus(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	respData := &ReqResp{}
	if err = json.Unmarshal(bts, respData); err != nil { // message, data
		return nil, err
	}
	return respData, nil
}

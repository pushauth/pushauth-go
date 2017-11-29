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
	SiteURL   = "https://api.pushauth.io"
	PushURL   = SiteURL + "/push/send"
	QRShowURL = SiteURL + "/qr/show"
	StatusURL = SiteURL + "/push/status"

	AppJSONMime = "application/json"
)

var (
	// HMAC Error
	ErrorHMACInvalid = errors.New("invalid hmac")

	// Encode error
	ErrorCannotEncode = errors.New("could not encode data")

	// Status Waiter Error
	ErrorStatusWaitClosed = errors.New("wait ended")

	// HTTP API Errors
	ErrorAccessDenied        = errors.New("access denied")
	ErrorNotFound            = errors.New("not found")
	ErrorMethodNotAllowed    = errors.New("method not allowed")
	ErrorUnprocessableEntity = errors.New("unprocessable entity")
	ErrorInternalServerError = errors.New("internal server error")
)

type AuthMode string

var (
	ModePush AuthMode = "push"
	ModeCode AuthMode = "code"
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
	return &PushAuth{publicKey: publicKey, privateKey: privateKey,
		hash: hmac.New(sha256.New, privateKey), WaitTime: waitTime}
}

func GetWaiterChans() (chan *StatusRespWait, chan struct{}) {
	return make(chan *StatusRespWait, 1), make(chan struct{}, 1)
}

func (p *PushAuth) getHMAC(data []byte) []byte {
	p.hash.Reset()
	p.hash.Write(data)
	return p.hash.Sum(nil)
}

func (p *PushAuth) encodeData(data interface{}) ([]byte, error) {
	bts, err := json.Marshal(data)

	if err != nil {
		return nil, ErrorCannotEncode
	}
	msg := encodeBase64(bts)
	reqData := &ReqData{PublicKey: string(p.publicKey),
		Data: fmt.Sprintf("%s.%s", msg, encodeBase64(p.getHMAC([]byte(msg))))}

	var marshaled []byte
	if marshaled, err = json.Marshal(reqData); err != nil {
		return nil, ErrorCannotEncode
	}
	return marshaled, nil
}

func (p *PushAuth) decodeData(data []byte, out interface{}) error {
	splits := bytes.Split(data, []byte("."))

	generatedHMAC := encodeBase64(p.getHMAC(splits[0]))

	if !hmac.Equal(splits[1], []byte(generatedHMAC)) {
		return ErrorHMACInvalid
	}

	decoded := decodeBase64(splits[0])

	if err := json.Unmarshal(decoded, out); err != nil {
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
	var (
		dPush []byte
		err   error
	)
	req := Req{Mode: ModePush, FlashResponse: flashResponse}
	dPlain := ReqSingle{AddrTo: to, Req: req}
	if dPush, err = p.encodeData(dPlain); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(dPush)

	return p.basicRequest(PushURL, AppJSONMime, reader, !flashResponse)
}

func (p *PushAuth) PushMult(to []string, flashResponse bool) (*ReqDataResp, error) {
	var (
		dPush []byte
		err   error
	)
	req := Req{Mode: ModePush, FlashResponse: flashResponse}
	mapTo := make(map[string]string)
	for idx := range to {
		mapTo[strconv.Itoa(idx+1)] = to[idx]
	}
	dPlain := ReqMultiple{AddrTo: mapTo, Req: req}
	if dPush, err = p.encodeData(dPlain); err != nil {
		return nil, err
	}

	return p.basicRequest(PushURL, AppJSONMime, bytes.NewReader(dPush), !flashResponse)
}

func (p *PushAuth) CodeSingle(to, code string) (*ReqDataResp, error) {
	var (
		dPush []byte
		err   error
	)
	req := Req{Mode: ModeCode, Code: code}
	dPlain := ReqSingle{AddrTo: to, Req: req}
	if dPush, err = p.encodeData(dPlain); err != nil {
		return nil, err
	}

	resp, err := p.basicRequest(PushURL, AppJSONMime, bytes.NewReader(dPush), false)
	if err != nil {
		return nil, err
	}

	resp.Answer = true
	return resp, nil
}

func (p *PushAuth) WaitForStatus(hash string, out chan<- *StatusRespWait, closer <-chan struct{}) {
	var (
		data []byte
		err  error
	)
	if data, err = p.encodeData(&CheckRequest{ReqHash: hash}); err != nil {
		out <- &StatusRespWait{nil, err}
		return
	}

	ticker := time.NewTicker(p.WaitTime)
	for {
		select {
		case <-ticker.C:
			resp, err := doPostRequest(StatusURL, AppJSONMime, bytes.NewReader(data))
			if err != nil {
				out <- &StatusRespWait{nil, err}
				return
			}
			status := &StatusResp{}
			if err := p.decodeData([]byte(resp.Data), status); err != nil {
				out <- &StatusRespWait{nil, err}
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
	var (
		bts []byte
	)
	resp, err := http.Post(url, contentType, data)

	if err != nil {
		return nil, err
	}

	if bts, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	if err = checkStatus(resp.StatusCode); err != nil {
		return nil, err
	}

	respData := &ReqResp{}
	if err = json.Unmarshal(bts, respData); err != nil { // message, data
		return nil, err
	}
	return respData, nil
}

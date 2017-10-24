package pushauth

import (
    "testing"
    "fmt"
    "time"
)

var (
    publicKey = []byte("G1Bvs3iKInFszyH1YfER33ZiS1wZl29t")
    privateKey = []byte("uzsDE0yR1OMuYquIA7QpxIWA7pfDzWej")
    email = "your@email.com"
)

func TestPushAuth_EncodeData(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
    mp := make(map[string]string)
    mp["addr_to"] = email
    mp["mode"] = "push"

    fmt.Println(pushAuth.encodeData(mp))
}


func TestPushAuth_SendPushSingle(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
    resp, err := pushAuth.PushSingle(email, false)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }
    
    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}

func TestPushAuth_SendPushSingleDoWait(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
    resp, err := pushAuth.PushSingle(email, true)
    if err != nil {
        t.Error("error is not nil: ", err)
    }

    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }

    out, closer := GetWaiterChans()
    go pushAuth.WaitForStatus(resp.ReqHash, out, closer)
    res := <- out

    if res.Error != nil {
        t.Error(res.Error)
        t.FailNow()
    }
}


func TestPushAuth_SendCodeSingle(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
    resp, err := pushAuth.CodeSingle(email, "12-12-12")
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }
    
    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}

// Fails, as server throws error
func TestPushAuth_SendPushMult(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
    resp, err := pushAuth.PushMult([]string{email, email}, false)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }

    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}
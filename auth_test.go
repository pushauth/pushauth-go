package pushauth

import (
    "testing"
    "fmt"
)

var (
    publicKey = []byte("G1Bvs3iKInFszyH1YfER33ZiS1wZl29t")
    privateKey = []byte("uzsDE0yR1OMuYquIA7QpxIWA7pfDzWej")
    email = "your@email.com"
)

func TestPushAuth_EncodeData(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    
    mp := make(map[string]string)
    
    mp["addr_to"] = email
    mp["mode"] = "push"
    
    fmt.Println(pushAuth.encodeData(mp))
}


func TestPushAuth_SendPushSingle(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.PushSingle(email, true, nil)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }
    
    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}


func TestPushAuth_SendPushSingleDoNotWait(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.PushSingle(email, false, nil)
    if err != nil {
        t.Error("error is not nil: ", err)
    }

    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}


func TestPushAuth_SendCodeSingle(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.CodeSingle(email, "12-12-12", nil)
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
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.PushMult([]string{email, email}, true, nil)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }

    if resp.ReqHash == "" {
        t.Error("empty request hash")
    }
}
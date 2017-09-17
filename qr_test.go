package pushauth

import (
    "log"
    "testing"
    "time"
)

func TestPushAuth_GetQR(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.GetQR(250)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }

    if resp.ReqHash == "" || resp.QRurl == "" {
        t.Error("empty request hash")
    }

    log.Println(resp.QRurl)

    out, closer := make(chan *StatusRespWait, 1), make(chan struct{}, 1)
    go pushAuth.WaitForStatus(resp.ReqHash, out, closer)
    waitResult := <- out

    if waitResult.Error != nil {
        t.Error("error when waiting is not nil: ", waitResult.Error)
        t.FailNow()
    }

    if waitResult.Answered != true {
        t.Error("answered != true")
    }
}

func TestPushAuth_GetQRTestCloser(t *testing.T) {
    pushAuth := NewPushAuth(publicKey, privateKey)
    resp, err := pushAuth.GetQR(250)
    if err != nil {
        t.Error("error is not nil: ", err)
        t.FailNow()
    }

    if resp.ReqHash == "" || resp.QRurl == "" {
        t.Error("empty request hash")
    }

    log.Println(resp.QRurl)

    out, closer := make(chan *StatusRespWait, 1), make(chan struct{}, 1)
    go pushAuth.WaitForStatus(resp.ReqHash, out, closer)

    <- time.NewTicker(5*time.Second).C
    closer <- struct{}{}
    waitResult := <- out

    if waitResult.Error != ErrorStatusWaitClosed {
        t.Error("error when waiting is not nil: ", waitResult.Error)
        t.FailNow()
    }
}
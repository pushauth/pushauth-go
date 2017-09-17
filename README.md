# pushauth-go
Simple Golang library for authenticating with PushAuth


## How to use?
All this samples are taken from tests, so they may include `testing` import

### Before we start
There were no extensive testing, so be warned!  

Push and Code requests are waiting internaly, not like QR request.  
As for now there is no way to cancel Push and Code wait.

Also, `WaitForStatus` checks response every 2 seconds, and this value is fixed for now.

### Simple push request
`PushSingle`, `PushMult` second argument is !`flash_response`. So if you want to set `flash_response` == true, then second argument should be `false`!

```
pushAuth := pushauth.NewPushAuth([]byte("publicKey"), []byte("privateKey"))
resp, err := pushAuth.PushSingle("client@email.com", true, nil)
if err != nil {
    // Some error
}

// Check if answered in resp.Answered
// Can check for answer(bool) in resp.Answer
if resp.Answered == false {
    // Not answered, maybe should wait?
}
```

### Simple code request
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey)
resp, err := pushAuth.CodeSingle(email, "12-12-12", nil)

if err != nil {
    // Some error
}

// Nothing to check, as if err == nil then no checkable answer is returned
```

### Simple push for multiple (in progress 'cause server fails)
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey)
resp, err := pushAuth.PushMult([]string{"client1@email.com", client2@email.com}, true, nil)
if err != nil {
    t.Error("error is not nil: ", err)
    t.FailNow()
}

// Check if answered in resp.Answered
// Can check for answer(bool) in resp.Answer
if resp.Answer == false {
    // Do something
}
```

### Get QR code 
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey)
resp, err := pushAuth.GetQR(250) // 250 is image size
if err != nil {
    t.Error("error is not nil: ", err)
    t.FailNow()
}

// resp.QRurl holds URL for QR code
if resp.ReqHash == "" || resp.QRurl == "" {
    t.Error("empty request hash")
}

// log.Println(resp.QRurl)

// And we wait for QR to be scanned
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
```

Or if you want to stop waiting after some timeout:
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey)
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

// Simulate waiting for 5 seconds, as if user clicks to stop waiting or something like that.
<- time.NewTicker(5*time.Second).C
closer <- struct{}{}
waitResult := <- out

if waitResult.Error != ErrorStatusWaitClosed {
    // Some error that you should handle
}
```

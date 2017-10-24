# pushauth-go
Simple Golang library for authenticating with PushAuth


## How to use?

### Before we start
*There were no extensive testing, so be warned!*

### New PushAuth instance
Create new instance by calling `NewPushAuth` method like so:
```
pushAuth := NewPushAuth(publicKey, privateKey, WaitTime)
```
Where `publicKey` and `privateKey` are bytes of public and private keys.
`WaitTime` is refresh interval when waiting for answer by hash. More on this in `WaitForStatus` section.
`WaitTime` can be changed later by changing `PushAuth`.`WaitTime`.

### WaitForStatus
`WaitForStatus` method is mandatory is `flashResponse` is set to `true` or using QR code.
`WaitForStatus` has three arguments:
1) `hash` - hash to check on. Returned by request
2) `out` - channel to send response to.
3) `closer` - channel to wait close signal from.

This method will check result for hash, wait duration setted in `PushAuth`.`WaitTime` parameter and then check again.
It will check for result until one of three happens:

- Wait will be closed
- Error is returned when request is sent
- Hash is answered and response code is 200

Basic usage should be like this:
```
out, closer := GetWaiterChans()
go pushAuth.WaitForStatus(resp.ReqHash, out, closer)
res := <- out
```

Or you can stop waiting, because it will wait infinitely for response:
```
out, closer := GetWaiterChans()
go pushAuth.WaitForStatus(resp.ReqHash, out, closer)

<- time.NewTicker(5*time.Second).C
close(closer)

waitResult := <-out
// waitResult.Error == ErrorStatusWaitClosed
```
When you stop waiting `StatusRespWait`.`Error` will be equal to `ErrorStatusWaitClosed`.

### Simple push request
`PushSingle` has two parameters: 
1) `email` - client to send push to
2) `flashResponse` - `flash_response` parameter. Set it to `false` if you want to wait for response.
If `flashResponse` == `true` - `req_hash` will be returned and you should wait for response manually.
Wait can be done by calling `WaitForStatus` method. Later on this.

```
pushAuth := NewPushAuth(publicKey, privateKey, 2*time.Second)
resp, err := pushAuth.PushSingle(email, false)
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
Code request returns immediately.
**Be aware that code request cannot be checked!** It can be only sent.
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey, 2*time.Second)
resp, err := pushAuth.CodeSingle(email, "12-12-12")

if err != nil {
    // Some error
}

// Nothing to check
```

### Simple push for multiple (in progress 'cause server fails)
```
pushAuth := pushauth.NewPushAuth(publicKey, privateKey, 2*time.Second)
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
QR request should be waited manually
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

// resp.QRurl holds URL for QR so it should be displayed to user somehow
// log.Println(resp.QRurl)

// And we wait for QR to be scanned
out, closer := GetWaiterChans()
go pushAuth.WaitForStatus(resp.ReqHash, out, closer)
waitResult := <-out

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

out, closer := GetWaiterChans()
go pushAuth.WaitForStatus(resp.ReqHash, out, closer)

// Simulate waiting for 5 seconds, as if user clicks to stop waiting or something like that.
<- time.NewTicker(5*time.Second).C
closer <- struct{}{}
// Or close(closer)
waitResult := <- out

if waitResult.Error != ErrorStatusWaitClosed {
    // Some error that you should handle
}
```

package pushauth

import (
    "bytes"
)

func (p *PushAuth) GetQR(size int) (*QRResponse, error) {
    dPlain := QRRequest{Image:QRImage{Size:size, BackgroundColor:"255,255,255",
    Color:"40,0,40", Margin:1}}
    dPush := p.encodeData(dPlain)
    
    reader := bytes.NewReader(dPush)

    respData, err := doPostRequest(QRShowURL, AppJSONMime, reader)
    if err != nil {
        return nil, err
    }

    respParsed := &QRResponse{}

    if err = p.decodeData([]byte(respData.Data), respParsed); err != nil {
        return nil, err
    }
    return respParsed, nil
}
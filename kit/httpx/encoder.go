package httpx

import (
	"encoding/json"
	stdhttp "net/http"

	"mall/kit/codeerror"
	"mall/kit/logger"
)

// CommonResp 统一 HTTP 响应包装
type CommonResp struct {
	Code     int               `json:"code"`
	Msg      string            `json:"msg"`
	MsgTrans map[string]string `json:"msg_trans"`
	Data     json.RawMessage   `json:"data"`
}

// ResponseEncoder 成功响应编码
func ResponseEncoder(w stdhttp.ResponseWriter, _ *stdhttp.Request, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusOK)
	bs, err := json.Marshal(v)
	if err != nil {
		logger.Errorf("ResponseEncoder marshal: %v", err)
		return codeerror.ErrorSystem
	}
	resp := CommonResp{
		Code: 0,
		Msg:  "success",
		Data: json.RawMessage(bs),
	}
	return json.NewEncoder(w).Encode(resp)
}

// ErrorEncoder 错误响应编码
func ErrorEncoder(w stdhttp.ResponseWriter, _ *stdhttp.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusOK)

	ce, ok := codeerror.As(err)
	if !ok {
		logger.Errorf("ErrorEncoder unknown error: %v", err)
		ce = codeerror.ErrorSystem
	}
	resp := CommonResp{
		Code: int(ce.Code),
		Msg:  ce.Msg,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

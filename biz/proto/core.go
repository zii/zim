package proto

import "fmt"

type Response struct {
	Code int         `json:"code" example:"200"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg" example:"success"`
}

type Success = Response
type Fail = Response

func (f *Fail) Error() string {
	return fmt.Sprintf("%d %s", f.Code, f.Msg)
}

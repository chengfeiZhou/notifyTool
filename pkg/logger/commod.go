package logger

import (
	"reflect"
)

var (
	ErrorParam            = AppError{code: 1001, msg: "Invalid Params"}
	ErrorEmptyData        = AppError{code: 1004, msg: "Data is empty"}
	ErrorMethodNotSupport = AppError{code: 1005, msg: "Method does not support"}
	ErrorMethod           = AppError{code: 1006, msg: "Method execution exception"}
	ErrorInstantiation    = AppError{code: 1007, msg: "Instance creation exception"}

	ErrorRouter          = AppError{code: 3001, msg: "Router execution exception"} // API Router
	ErrorRouterRegister  = AppError{code: 3002, msg: "Router Register exception"}
	ErrorAgentStart      = AppError{code: 3003, msg: "Agent Start exception"}
	ErrorMakeMiddleware  = AppError{code: 3004, msg: "Make Middleware exception"}
	ErrorHandleError     = AppError{code: 3005, msg: "Error Handle exception"}
	ErrorRequestExecutor = AppError{code: 3006, msg: "HTTP Request Executor exception"}
	ErrorServerPlugin    = AppError{code: 3007, msg: "ServerPlugin exception"}

	ErrorParamsIncomplete     = AppError{code: 4001, msg: "Incomplete parameters"} // 参数
	ErrorAuthParamsIncomplete = AppError{code: 4101, msg: "Register/UnRegister body is null"}

	ErrorHost          = AppError{code: 5001, msg: "Request address exception"} // 网络请求错误
	ErrorHTTPHandle    = AppError{code: 5002, msg: "HTTP Handle exception"}
	ErrorNetForwarding = AppError{code: 5003, msg: "Network forwarding error"}
	ErrorMiddleware    = AppError{code: 5004, msg: "Middleware Handle error"}

	// 可扩展...
)

// AppError 业务逻辑错误对象
type AppError struct {
	msg  string
	code int
}

// Code 返回错误代码
func (r AppError) Code() int {
	return r.code
}

// Error 返回错误信息
func (r AppError) Error() string {
	return r.msg
}

// IsAppError 比较
func IsAppError(err error) bool {
	return reflect.TypeOf(&err) == reflect.TypeOf((*AppError)(nil))
}

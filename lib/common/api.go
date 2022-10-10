package common

const (
	DataTypeApi DataType = "api"
)

type ApiRequestData struct {
	DataType DataType `json:"data_type"`
	Package  string   `json:"package"`
	Plugin   string   `json:"plugin"`
	Args     []string `json:"args"`
}

func NewApiRequestData(Package string, plugin string, args []string) ApiRequestData {
	return ApiRequestData{
		DataType: DataTypeApi,
		Package:  Package,
		Plugin:   plugin,
		Args:     args,
	}
}

type ApiResultData struct {
	DataType DataType `json:"data_type"`
	Message  string   `json:"message"`
	Code     int      `json:"code"`
}

func NewApiResultData(message string, code int) ApiResultData {
	return ApiResultData{
		DataType: DataTypeApi,
		Message:  message,
		Code:     code,
	}
}

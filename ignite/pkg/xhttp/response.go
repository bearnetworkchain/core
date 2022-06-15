package xhttp

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ResponseJSON 使用 status 作為 http 狀態和數據向 w 寫入 JSON 響應
// 作為有效載荷。
func ResponseJSON(w http.ResponseWriter, status int, data interface{}) error {
	bdata, err := json.Marshal(data)
	if err != nil {
		status = http.StatusInternalServerError
		bdata, _ = json.Marshal(NewErrorResponse(errors.New(http.StatusText(status))))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bdata)
	return err
}

// ErrorResponseBody 是應該發送到的錯誤消息的骨架
// 客戶。
type ErrorResponseBody struct {
	Error ErrorResponse `json:"error"`
}

// ErrorResponse 保存錯誤消息。
type ErrorResponse struct {
	Message string `json:"message"`
}

// NewErrorResponse 從 err 創建一個新的 http 錯誤響應。
func NewErrorResponse(err error) ErrorResponseBody {
	return ErrorResponseBody{
		Error: ErrorResponse{
			Message: err.Error(),
		},
	}
}

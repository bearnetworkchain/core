package networktypes

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrInvalidRequest 是在處理無效請求的方法中返回的錯誤
type ErrInvalidRequest struct {
	requestID uint64
}

// 錯誤實現錯誤
func (err ErrInvalidRequest) Error() string {
	return fmt.Sprintf("要求 %d 是無效的", err.requestID)
}

// NewWrappedErrInvalidRequest返回一個包裝好的 ErrInvalidRequest
func NewWrappedErrInvalidRequest(requestID uint64, message string) error {
	return errors.Wrap(ErrInvalidRequest{requestID: requestID}, message)
}

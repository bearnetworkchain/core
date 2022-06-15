package cosmoserror

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal       = errors.New("內部錯誤")
	ErrInvalidRequest = errors.New("無效的請求")
	ErrNotFound       = errors.New("未找到")
)

func Unwrap(err error) error {
	s, ok := status.FromError(err)
	if ok {
		switch s.Code() {
		case codes.NotFound:
			return ErrNotFound
		case codes.InvalidArgument:
			return ErrInvalidRequest
		case codes.Internal:
			return ErrInternal
		}
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		return unwrapped
	}
	return err
}

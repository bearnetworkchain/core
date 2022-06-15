// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tendermint/spn/x/reward/types"
	"google.golang.org/grpc"
)

// RewardClient是 RewardClient 類型的自動生成的模擬類型
type RewardClient struct {
	mock.Mock
}

// Params 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *RewardClient) Params(ctx context.Context, in *types.QueryParamsRequest, opts ...grpc.CallOption) (*types.QueryParamsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryParamsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryParamsRequest, ...grpc.CallOption) *types.QueryParamsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryParamsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryParamsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RewardPool提供具有給定字段的模擬函數：ctx、in、opts
func (_m *RewardClient) RewardPool(ctx context.Context, in *types.QueryGetRewardPoolRequest, opts ...grpc.CallOption) (*types.QueryGetRewardPoolResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryGetRewardPoolResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryGetRewardPoolRequest, ...grpc.CallOption) *types.QueryGetRewardPoolResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryGetRewardPoolResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryGetRewardPoolRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RewardPoolAll提供具有給定字段的模擬函數：ctx、in、opts
func (_m *RewardClient) RewardPoolAll(ctx context.Context, in *types.QueryAllRewardPoolRequest, opts ...grpc.CallOption) (*types.QueryAllRewardPoolResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryAllRewardPoolResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryAllRewardPoolRequest, ...grpc.CallOption) *types.QueryAllRewardPoolResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryAllRewardPoolResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryAllRewardPoolRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewRewardClientT interface {
	mock.TestingT
	Cleanup(func())
}

// NewRewardClient 創建一個新的 RewardClient 實例。它還在模擬上註冊了一個測試接口和一個清理函數來斷言模擬期望。
func NewRewardClient(t NewRewardClientT) *RewardClient {
	mock := &RewardClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

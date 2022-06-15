// Code generated by mockery v2.11.0. DO NOT EDIT.

package mocks

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// StakingClient 是 StakingClient 類型的自動生成的模擬類型
type StakingClient struct {
	mock.Mock
}

// Delegation 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Delegation(ctx context.Context, in *types.QueryDelegationRequest, opts ...grpc.CallOption) (*types.QueryDelegationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryDelegationResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryDelegationRequest, ...grpc.CallOption) *types.QueryDelegationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryDelegationResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryDelegationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DelegatorDelegations 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) DelegatorDelegations(ctx context.Context, in *types.QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*types.QueryDelegatorDelegationsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryDelegatorDelegationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryDelegatorDelegationsRequest, ...grpc.CallOption) *types.QueryDelegatorDelegationsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryDelegatorDelegationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryDelegatorDelegationsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DelegatorUnbondingDelegations 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) DelegatorUnbondingDelegations(ctx context.Context, in *types.QueryDelegatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*types.QueryDelegatorUnbondingDelegationsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryDelegatorUnbondingDelegationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryDelegatorUnbondingDelegationsRequest, ...grpc.CallOption) *types.QueryDelegatorUnbondingDelegationsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryDelegatorUnbondingDelegationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryDelegatorUnbondingDelegationsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DelegatorValidator 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) DelegatorValidator(ctx context.Context, in *types.QueryDelegatorValidatorRequest, opts ...grpc.CallOption) (*types.QueryDelegatorValidatorResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryDelegatorValidatorResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryDelegatorValidatorRequest, ...grpc.CallOption) *types.QueryDelegatorValidatorResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryDelegatorValidatorResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryDelegatorValidatorRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DelegatorValidators 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) DelegatorValidators(ctx context.Context, in *types.QueryDelegatorValidatorsRequest, opts ...grpc.CallOption) (*types.QueryDelegatorValidatorsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryDelegatorValidatorsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryDelegatorValidatorsRequest, ...grpc.CallOption) *types.QueryDelegatorValidatorsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryDelegatorValidatorsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryDelegatorValidatorsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HistoricalInfo 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) HistoricalInfo(ctx context.Context, in *types.QueryHistoricalInfoRequest, opts ...grpc.CallOption) (*types.QueryHistoricalInfoResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryHistoricalInfoResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryHistoricalInfoRequest, ...grpc.CallOption) *types.QueryHistoricalInfoResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryHistoricalInfoResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryHistoricalInfoRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Params 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Params(ctx context.Context, in *types.QueryParamsRequest, opts ...grpc.CallOption) (*types.QueryParamsResponse, error) {
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

// Pool 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Pool(ctx context.Context, in *types.QueryPoolRequest, opts ...grpc.CallOption) (*types.QueryPoolResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryPoolResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryPoolRequest, ...grpc.CallOption) *types.QueryPoolResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryPoolResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryPoolRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Redelegations 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Redelegations(ctx context.Context, in *types.QueryRedelegationsRequest, opts ...grpc.CallOption) (*types.QueryRedelegationsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryRedelegationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryRedelegationsRequest, ...grpc.CallOption) *types.QueryRedelegationsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryRedelegationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryRedelegationsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnbondingDelegation提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) UnbondingDelegation(ctx context.Context, in *types.QueryUnbondingDelegationRequest, opts ...grpc.CallOption) (*types.QueryUnbondingDelegationResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryUnbondingDelegationResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryUnbondingDelegationRequest, ...grpc.CallOption) *types.QueryUnbondingDelegationResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryUnbondingDelegationResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryUnbondingDelegationRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Validator提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Validator(ctx context.Context, in *types.QueryValidatorRequest, opts ...grpc.CallOption) (*types.QueryValidatorResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryValidatorResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryValidatorRequest, ...grpc.CallOption) *types.QueryValidatorResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryValidatorResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryValidatorRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidatorDelegations提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) ValidatorDelegations(ctx context.Context, in *types.QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*types.QueryValidatorDelegationsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryValidatorDelegationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryValidatorDelegationsRequest, ...grpc.CallOption) *types.QueryValidatorDelegationsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryValidatorDelegationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryValidatorDelegationsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidatorUnbondingDelegations提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) ValidatorUnbondingDelegations(ctx context.Context, in *types.QueryValidatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*types.QueryValidatorUnbondingDelegationsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryValidatorUnbondingDelegationsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryValidatorUnbondingDelegationsRequest, ...grpc.CallOption) *types.QueryValidatorUnbondingDelegationsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryValidatorUnbondingDelegationsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryValidatorUnbondingDelegationsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Validators 提供具有給定字段的模擬函數：ctx、in、opts
func (_m *StakingClient) Validators(ctx context.Context, in *types.QueryValidatorsRequest, opts ...grpc.CallOption) (*types.QueryValidatorsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *types.QueryValidatorsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.QueryValidatorsRequest, ...grpc.CallOption) *types.QueryValidatorsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.QueryValidatorsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.QueryValidatorsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewQueryClient 創建一個新的 StakingClient 實例。它還註冊了一個清理函數來斷言模擬期望。
func NewQueryClient(t testing.TB) *StakingClient {
	mock := &StakingClient{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

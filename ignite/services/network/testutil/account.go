package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
)

const (
	TestAccountName = "test"
)

//NewTestAccount 使用內存密鑰環後端創建一個用於測試目的的帳戶
func NewTestAccount(t *testing.T, name string) cosmosaccount.Account {
	r, err := cosmosaccount.NewInMemory()
	assert.NoError(t, err)
	account, _, err := r.Create(name)
	assert.NoError(t, err)
	return account
}

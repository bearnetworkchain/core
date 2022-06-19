package entrywriter_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/entrywriter"
)

type WriterWithError struct{}

func (WriterWithError) Write(_ []byte) (n int, err error) {
	return 0, errors.New("writer with error")
}

func TestWrite(t *testing.T) {
	header := []string{"foobar", "bar", "foo"}

	entries := [][]string{
		{"foo", "bar", "foobar"},
		{"bar", "foobar", "foo"},
		{"foobar", "foo", "bar"},
	}

	require.NoError(t, entrywriter.Write(io.Discard, header, entries...))
	require.NoError(t, entrywriter.Write(io.Discard, header), "應該不允許進入")

	err := entrywriter.Write(io.Discard, []string{})
	require.ErrorIs(t, err, entrywriter.ErrInvalidFormat, "應該防止沒有標題")

	entries[0] = []string{"foo", "bar"}
	err = entrywriter.Write(io.Discard, header, entries...)
	require.ErrorIs(t, err, entrywriter.ErrInvalidFormat, "應防止條目長度不符合")

	var wErr WriterWithError
	require.Error(t, entrywriter.Write(wErr, header, entries...), "應該捕捉作家錯誤")
}

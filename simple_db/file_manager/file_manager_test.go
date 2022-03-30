package file_manager

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestFileManager(t *testing.T) {
	// require.Equal(t, 1, 2) //确保用例能执行
    fm, _ := NewFileManager("file_test", 400)
    
	blk := NewBlockId("testfile", 2)
	p1 := NewPageBySize(fm.BlockSize())
	pos1 := uint64(88)
	s := "abcdefghijklm"
	p1.SetString(pos1, s)
	size := p1.MaxLengthForString(s)
	pos2 := pos1 + size 
	val := uint64(345)
	p1.SetInt(pos2, val)
    fm.Write(blk, p1)

	p2 := NewPageBySize(fm.BlockSize())
	fm.Read(blk, p2)
	
	require.Equal(t, val, p2.GetInt(pos2))

	require.Equal(t, s, p2.GetString(pos1))
}
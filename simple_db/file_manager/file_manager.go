package file_manager

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileManager struct {
	db_directory string
	block_size   uint64
	is_new       bool
	open_files   map[string]*os.File
	mu           sync.Mutex
}

func NewFileManager(db_directory string, block_size uint64) (*FileManager, error) {
	file_manager := FileManager{
		db_directory: db_directory,
		block_size:   block_size,
		is_new:       false,
		open_files:   make(map[string]*os.File),
	}

	if _, err := os.Stat(db_directory); os.IsNotExist(err) {
		//目录不存在则创建
		file_manager.is_new = true
		err = os.Mkdir(db_directory, os.ModeDir)
		if err != nil {
			return nil, err
		}
	} else {
		//目录存在，则先清楚目录下的临时文件
		err := filepath.Walk(db_directory, func(path string, info os.FileInfo, err error) error {
			mode := info.Mode()
			if mode.IsRegular() {
				name := info.Name()
				if strings.HasPrefix(name, "temp") {
					//删除临时文件
					os.Remove(filepath.Join(path, name))
				}
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return &file_manager, nil
}

func (f *FileManager) getFile(file_name string) (*os.File, error) {
	path := filepath.Join(f.db_directory, file_name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	f.open_files[path] = file

	return file, nil
}

func (f *FileManager) Read(blk *BlockId, p *Page) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := f.getFile(blk.FileName())
	if err != nil {
		return 0, err
	}
	defer file.Close()
	count, err := file.ReadAt(p.contents(), int64(blk.Number()*f.block_size))
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (f FileManager) Write(blk *BlockId, p *Page) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := f.getFile(blk.FileName())
	if err != nil {
		return 0, err
	}
	defer file.Close()

	n, err := file.WriteAt(p.contents(), int64(blk.Number()*f.block_size))
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (f *FileManager) size(file_name string) (uint64, error) {
	file, err := f.getFile(file_name)
	if err != nil {
		return 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return uint64(fi.Size()) / f.block_size, nil

}

func (f *FileManager) Append(file_name string) (BlockId, error) {
	new_block_num, err := f.size(file_name)
	if err != nil {
		return BlockId{}, err
	}

	blk := NewBlockId(file_name, new_block_num)
	file, err := f.getFile(blk.FileName())
	if err != nil {
		return BlockId{}, err
	}

	b := make([]byte, f.block_size)
	_, err = file.WriteAt(b, int64(blk.Number()*f.block_size)) //读入空数据相当于扩大文件长度
	if err != nil {
		return BlockId{}, nil
	}

	return *blk, nil
}

func (f *FileManager) IsNew() bool {
	return f.is_new
}

func (f *FileManager) BlockSize() uint64 {
	return f.block_size
}

package util

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type MMap struct {
	Mapper
}

func (m *MMap) MapFromFile(f *os.File) ([]byte, error) {
	return syncFromFile(f)
}

func (m *MMap) MapToFile(data []byte, f *os.File) ([]byte, error) {
	return syncToFile(data, f)
}

func (m *MMap) Unmap(data []byte) error {
	return unix.Munmap(data)
}

func syncToFile(data []byte, f *os.File) ([]byte, error) {
	// check target file readable/writable
	if f == nil {
		return nil, fmt.Errorf("file is not valid")
	}

	st, serr := f.Stat()
	if serr != nil {
		return nil, fmt.Errorf("input os.File stat has error: %s", serr)
	}
	if st.Mode().Perm()&0600 != 0600 {
		return nil, fmt.Errorf("file is not writable")
	}

	// check mapping data
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to map")
	}

	mapSize := len(data)

	err := unix.Ftruncate(int(f.Fd()), int64(mapSize))
	if err != nil {
		return nil, fmt.Errorf("failed to truncate the file, err: %s", err)
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to rewind the file pointer to the beginning, err: %s", err)
	}

	// remap
	d, err := unix.Mmap(int(f.Fd()), 0, int(mapSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("failed to mmap to file, err: %s", err)
	}
	_ = copy(d, data)
	err = unix.Msync(d, unix.MS_SYNC|unix.MS_INVALIDATE)
	if err != nil {
		return nil, fmt.Errorf("failed to sync mmap data, err: %s", err)
	}
	return d, nil
}

func syncFromFile(f *os.File) ([]byte, error) {
	// check target file readable
	if f == nil {
		return nil, fmt.Errorf("file is not valid")
	}

	st, serr := f.Stat()
	if serr != nil {
		return nil, fmt.Errorf("input os.File stat has error: %s", serr)
	}
	if st.Mode().Perm()&0400 != 0400 {
		return nil, fmt.Errorf("file is not readable")
	}

	fileLen := st.Size()
	d, err := unix.Mmap(int(f.Fd()), 0, int(fileLen), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("faild to mmap from file, err: %s", err)
	}
	return d, nil
}

type Mapper interface {
	MapFromFile(*os.File) ([]byte, error)
	MapToFile([]byte, *os.File) ([]byte, error)
	Unmap() error
}

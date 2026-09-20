package Api

import (
	"archive/tar"
	"errors"
)

const (
	S_IFMT   = 0170000
	S_IFREG  = 0100000
	S_IFDIR  = 0040000
	S_IFLNK  = 0120000
	S_IFCHR  = 0020000
	S_IFBLK  = 0060000
	S_IFIFO  = 0010000
	S_IFSOCK = 0140000
)

func UnpackMode(mode uint16) (byte, error) {
	fileType := mode & S_IFMT

	switch fileType {
	case S_IFREG:
		return tar.TypeReg, nil
	case S_IFDIR:
		return tar.TypeDir, nil
	case S_IFLNK:
		return tar.TypeSymlink, nil
	case S_IFCHR:
		return tar.TypeChar, nil
	case S_IFBLK:
		return tar.TypeBlock, nil
	case S_IFIFO:
		return tar.TypeFifo, nil
	case S_IFSOCK:
		return 0, errors.ErrUnsupported
	default:
		return tar.TypeReg, nil
	}
}

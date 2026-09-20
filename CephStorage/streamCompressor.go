package CephStorage

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	cephfs "github.com/ceph/go-ceph/cephfs"
)

var tempRootDir string

type CephMount struct {
	CephMount *cephfs.MountInfo
}

type Storage interface {
	Read(path string) (io.ReadCloser, error)
	Write(path string, data []byte) (io.WriteCloser, error)
}

func (cm *CephMount) CephRead(path string) (io.ReadCloser, error) {
	return cm.CephMount.Open(path, 0777, uint32(os.O_RDONLY))
}

func (cm *CephMount) CephWrite(path string) (io.WriteCloser, error) {
	if cm.CephMount == nil {
		return nil, errors.New("ceph mount is nil")
	}
	return cm.CephMount.Open(path, 777, uint32(os.O_WRONLY|os.O_CREATE|os.O_TRUNC))
}

func tarTypeFlag(mode uint16) (byte, error) {
	const (
		sIfmt   = 0170000
		sIfreg  = 0100000
		sIfdir  = 0040000
		sIflnk  = 0120000
		sIfchr  = 0020000
		sIfblk  = 0060000
		sIfifo  = 0010000
		sIfsock = 0140000
	)

	fileType := mode & sIfmt
	switch fileType {
	case sIfreg:
		return tar.TypeReg, nil
	case sIfdir:
		return tar.TypeDir, nil
	case sIflnk:
		return tar.TypeSymlink, nil
	case sIfchr:
		return tar.TypeChar, nil
	case sIfblk:
		return tar.TypeBlock, nil
	case sIfifo:
		return tar.TypeFifo, nil
	case sIfsock:
		return 0, errors.ErrUnsupported
	default:
		return tar.TypeReg, nil
	}
}

func (cm *CephMount) CompressSubvolume(subvolPath string, username string) error {
	workplaceMount := path.Dir(subvolPath)

	archvDir := path.Join(subvolPath, "arch")
	tarPath := path.Join(archvDir, "archv.tar")
	errMd := cm.CephMount.MakeDir(archvDir, uint32(755))
	if errMd != nil {
		println("error on creating tarpath dir: ",archvDir,"with error: ", errMd.Error())
	}

	writer, errWrite := cm.CephWrite(tarPath)
	println("In compress, values are: ", subvolPath, workplaceMount)
	if errWrite != nil {
		println("cephWrite error: ", errWrite.Error())
		return errors.New("Assigning a write object for the compressor failed")
	}
	println("entered tarify?")

	wrt := tar.NewWriter(writer)
	err := cm.tarifyBfs(subvolPath, wrt)
	if err != nil {
		println("Error in tarify top level", err.Error())
		return err
	}
	if err := wrt.Close(); err != nil {
		return fmt.Errorf("close tar writer: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close archive writer: %w", err)
	}

	cm.HardlinkBlob(tarPath, username)
	println("left tarify")
	return nil
}

// ended up being very weird, if a dir is met it skips to bottom level but it keeps BFS mechanics for normal files, it works so it does not matter
func (cm *CephMount) tarifyBfs(rootPath string, tarW *tar.Writer) error {
	println("did compress call")
	if cm.CephMount == nil {
		return fmt.Errorf("ceph mount is nil")
	}
	if rootPath == "" {
		return fmt.Errorf("root path is empty")
	}
	if tarW == nil {
		return fmt.Errorf("tar writer is nil")
	}

	var walk func(string) error
	walk = func(currentPath string) error {
		println("rec call")

		println("Opening directory: ", currentPath)
		dir, err := cm.CephMount.OpenDir(currentPath)
		if err != nil {
			return fmt.Errorf("open dir %q: %w", currentPath, err)
		}
		defer dir.Close()

		
		
		dir.RewindDir()
		for {
			entry, err := dir.ReadDir()
			if err != nil {
				return fmt.Errorf("read dir %q: %w", currentPath, err)
			}
			if entry == nil {
				break
			}

			if entry.Name() == "." || entry.Name() == ".."{
				continue
			}
			childPath := path.Join(currentPath, entry.Name())
			println("childPath: ", childPath)

			attrs, err := cm.CephMount.Statx(childPath, cephfs.StatxAllStats, cephfs.AtStatxDontSync)
			if err != nil {
				return fmt.Errorf("stat %q: %w", childPath, err)
			}
			mtime := time.Unix(attrs.Mtime.Sec, attrs.Mtime.Nsec)

			if entry.DType() == cephfs.DTypeDir {
				header := &tar.Header{
					Name:     childPath,
					Mode:     int64(attrs.Mode),
					Typeflag: tar.TypeDir,
					ModTime:  mtime,
				}
				if err := tarW.WriteHeader(header); err != nil {
					return fmt.Errorf("write dir header for %q: %w", childPath, err)
				}

				if err := walk(childPath); err != nil {
					return err
				}

				continue
			}

			file, err := cm.CephMount.Open(childPath, os.O_RDONLY, 0)
			if err != nil {
				return fmt.Errorf("open file %q: %w", childPath, err)
			}

			typeFlag, err := tarTypeFlag(attrs.Mode)
			if err != nil {
				file.Close()
				return fmt.Errorf("classify file type for %q: %w", childPath, err)
			}

			header := &tar.Header{
				Name:     childPath,
				Mode:     int64(attrs.Mode),
				Size:     int64(attrs.Size),
				Typeflag: typeFlag,
				ModTime:  mtime,
			}
			if err := tarW.WriteHeader(header); err != nil {
				file.Close()
				return fmt.Errorf("write file header for %q: %w", childPath, err)
			}
			file.Seek(0,io.SeekStart)

			if btsWrtn, err := io.CopyN(tarW, file, int64(attrs.Size)); err != nil {
				println("Copying")
				println("Bytes written: ", btsWrtn, "Bytes from stat: ", attrs.Size)
				file.Close()
				return fmt.Errorf("copy file contents for %q: %w", childPath, err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close file %q: %w", childPath, err)
			}
		}

		return nil
	}

	return walk(rootPath)
}

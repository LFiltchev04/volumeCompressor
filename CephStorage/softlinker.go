package CephStorage

import (
	"compressor/Configuration"
	"fmt"
	"os"
	"path"

	"github.com/opencontainers/go-digest"
)

func MakeSoftLink(hash string, crrlocation string) error {
	dgst, err := digest.Parse(hash)
	if err != nil {
		return fmt.Errorf("parse blob digest: %w", err)
	}

	relativePath, err := PathFor(blobPathSpec{
		digest: dgst,
	})
	if err != nil {
		return fmt.Errorf("build blob path: %w", err)
	}

	linkPath := path.Join(Configuration.Global.BaseConf.DistributionPV, relativePath)
	mount := &CephMount{CephMount: &Configuration.Global.GlobalMnt}
	if err := mount.ensureDir(path.Dir(linkPath)); err != nil {
		return fmt.Errorf("create blob directory: %w", err)
	}

	if err := mount.CephMount.Symlink(crrlocation, linkPath); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("create blob symlink: %w", err)
	}

	return nil
}

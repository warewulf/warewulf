package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containers/image/v5/types"
)

const (
	defaultCachePath = "/var/warewulf/vnfs-cache/oci/"
	blobPrefix       = "blobs"
	rootfsPrefix     = "rootfs"
)

type CacheOpt func(*Cache) error

func OptSetCachePath(path string) CacheOpt {
	return func(s *Cache) error {
		s.path = path
		return nil
	}
}

type Cache struct {
	path string
}

func (c *Cache) rootfsDir() string {
	return filepath.Join(c.path, rootfsPrefix)
}

func (c *Cache) blobDir() string {
	return filepath.Join(c.path, blobPrefix)
}

// checkEntry maps the given id a path within the cache. It will return an os.ErrNotExist if the entry path is empty.
func (c *Cache) checkEntry(id string) (string, error) {
	path := filepath.Join(c.rootfsDir(), id)
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("invalid entry %q is not a directory", path)
	}

	return path, nil
}

func (c *Cache) createEntry(id string) (string, error) {
	path := filepath.Join(c.rootfsDir(), id)
	if err := os.MkdirAll(path, 700); err != nil {
		return "", err
	}

	return path, nil
}

func NewCache(opts ...CacheOpt) (*Cache, error) {
	s := &Cache{
		path: filepath.Join(defaultCachePath),
	}

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(s.path, 0700); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Cache) Pull(ctx context.Context, uri string, sysCtx *types.SystemContext) (string, error) {
	p, err := newPuller(
		optSetBlobCachePath(c.blobDir()),
		optSetSystemContext(sysCtx),
	)
	if err != nil {
		return "", err
	}

	id, err := p.generateID(ctx, uri)
	if err != nil {
		return "", err
	}

	// look up cache entry and return if it exists
	path, err := c.checkEntry(id)
	if err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// cache entry does not exist so we must create it
	path, err = c.createEntry(id)
	if err != nil {
		return "", err
	}

	// populate entry
	if err := p.pull(ctx, uri, path); err != nil {
		// clean up entry on error
		os.RemoveAll(path)
		return "", err
	}

	return path, nil
}

package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/service/playground_attachment/driver"
)

type Driver struct {
	basePath string
}

func New(basePath string) *Driver {
	if strings.TrimSpace(basePath) == "" {
		basePath = "./data/playground-attachments"
	}
	return &Driver{basePath: basePath}
}

func (d *Driver) Name() string {
	return "local"
}

func (d *Driver) Put(_ context.Context, objectKey string, _ string, body io.Reader) (int64, error) {
	targetPath, err := d.objectPath(objectKey)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return io.Copy(file, body)
}

func (d *Driver) Open(_ context.Context, objectKey string) (io.ReadCloser, error) {
	targetPath, err := d.objectPath(objectKey)
	if err != nil {
		return nil, err
	}
	return os.Open(targetPath)
}

func (d *Driver) Delete(_ context.Context, objectKey string) error {
	targetPath, err := d.objectPath(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (d *Driver) Reference(_ context.Context, _ string, _ time.Time) (*driver.Reference, error) {
	return nil, driver.ErrReferenceUnsupported
}

func (d *Driver) objectPath(objectKey string) (string, error) {
	cleanKey := filepath.Clean(filepath.FromSlash(objectKey))
	if cleanKey == "." || filepath.IsAbs(cleanKey) || strings.HasPrefix(cleanKey, ".."+string(filepath.Separator)) || cleanKey == ".." {
		return "", fmt.Errorf("invalid object key")
	}

	baseAbs, err := filepath.Abs(d.basePath)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanKey))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid object key")
	}
	return targetAbs, nil
}

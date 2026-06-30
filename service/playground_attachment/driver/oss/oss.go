package oss

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/service/playground_attachment/driver"
)

type Config struct {
	Endpoint        string
	Bucket          string
	Region          string
	AccessKeyID     string
	AccessKeySecret string
	ObjectPrefix    string
}

type Driver struct {
	config Config
	client *http.Client
}

func New(config Config) *Driver {
	config.Endpoint = strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKeyID = strings.TrimSpace(config.AccessKeyID)
	config.AccessKeySecret = strings.TrimSpace(config.AccessKeySecret)
	config.ObjectPrefix = strings.Trim(strings.TrimSpace(config.ObjectPrefix), "/")
	return &Driver{config: config, client: http.DefaultClient}
}

func (d *Driver) Name() string {
	return "oss"
}

func (d *Driver) Put(ctx context.Context, objectKey string, contentType string, body io.Reader) (int64, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, err
	}
	req, err := d.newSignedRequest(ctx, http.MethodPut, objectKey, contentType, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.ContentLength = int64(len(data))
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("oss put failed with status %d", resp.StatusCode)
	}
	return int64(len(data)), nil
}

func (d *Driver) Open(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	req, err := d.newSignedRequest(ctx, http.MethodGet, objectKey, "", nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, os.ErrNotExist
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("oss get failed with status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (d *Driver) Delete(ctx context.Context, objectKey string) error {
	req, err := d.newSignedRequest(ctx, http.MethodDelete, objectKey, "", nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("oss delete failed with status %d", resp.StatusCode)
	}
	return nil
}

func (d *Driver) Reference(_ context.Context, objectKey string, expiresAt time.Time) (*driver.Reference, error) {
	if err := d.validateConfig(); err != nil {
		return nil, err
	}
	expires := expiresAt.Unix()
	fullKey := d.fullObjectKey(objectKey)
	resource := d.canonicalResource(fullKey)
	stringToSign := fmt.Sprintf("GET\n\n\n%d\n%s", expires, resource)
	signature := d.sign(stringToSign)

	objectURL, err := d.objectURL(fullKey)
	if err != nil {
		return nil, err
	}
	query := objectURL.Query()
	query.Set("OSSAccessKeyId", d.config.AccessKeyID)
	query.Set("Expires", fmt.Sprintf("%d", expires))
	query.Set("Signature", signature)
	objectURL.RawQuery = query.Encode()

	return &driver.Reference{URL: objectURL.String(), ExpiresAt: expires}, nil
}

func (d *Driver) newSignedRequest(ctx context.Context, method string, objectKey string, contentType string, body io.Reader) (*http.Request, error) {
	if err := d.validateConfig(); err != nil {
		return nil, err
	}
	fullKey := d.fullObjectKey(objectKey)
	objectURL, err := d.objectURL(fullKey)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, objectURL.String(), body)
	if err != nil {
		return nil, err
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	stringToSign := fmt.Sprintf("%s\n\n%s\n%s\n%s", method, contentType, date, d.canonicalResource(fullKey))
	req.Header.Set("Authorization", "OSS "+d.config.AccessKeyID+":"+d.sign(stringToSign))
	return req, nil
}

func (d *Driver) validateConfig() error {
	if d.config.Endpoint == "" || d.config.Bucket == "" || d.config.AccessKeyID == "" || d.config.AccessKeySecret == "" {
		return fmt.Errorf("oss driver is not configured")
	}
	return nil
}

func (d *Driver) fullObjectKey(objectKey string) string {
	cleanKey := strings.TrimLeft(objectKey, "/")
	if d.config.ObjectPrefix == "" {
		return cleanKey
	}
	return d.config.ObjectPrefix + "/" + cleanKey
}

func (d *Driver) canonicalResource(fullKey string) string {
	return "/" + d.config.Bucket + "/" + fullKey
}

func (d *Driver) objectURL(fullKey string) (*url.URL, error) {
	endpoint := d.config.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(parsed.Host, d.config.Bucket+".") {
		parsed.Host = d.config.Bucket + "." + parsed.Host
	}
	parsed.Path = "/" + escapeObjectKey(fullKey)
	parsed.RawQuery = ""
	return parsed, nil
}

func (d *Driver) sign(data string) string {
	mac := hmac.New(sha1.New, []byte(d.config.AccessKeySecret))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func escapeObjectKey(objectKey string) string {
	segments := strings.Split(objectKey, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

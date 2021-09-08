package gcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/94peter/sterna/util"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type Perm string

const (
	PermPublic  = Perm("public")
	PermPrivate = Perm("private")
	PermTmp     = Perm("tmp")
)

type Storage interface {
	GetAttr(ctx context.Context, key string, pm Perm) (*storage.ObjectAttrs, error)
	RemoveObject(ctx context.Context, key string, pm Perm) error
	GetPublicUrl(ctx context.Context, object string) (myurl string, err error)
	WriteString(ctx context.Context, key string, content string, pm Perm) (path string, err error)
	Write(ctx context.Context, key string, pm Perm, writeData func(w io.Writer) error) error
	OpenFile(ctx context.Context, key string, pm Perm) (io.Reader, error)
	SignedURL(key string, contentType string, pm Perm, expDuration time.Duration) (url string, err error)
}

type GcpConf struct {
	CredentialsFile string `yaml:"credentailsFile"`
	Bucket          string
	PublicBucket    string `yaml:"publicBucket"`
	TmpBucket       string `yaml:"tmpBucket"`

	credentials *google.Credentials
}

func (gcp *GcpConf) getCredentials(ctx context.Context) (*google.Credentials, error) {
	if gcp.credentials != nil {
		return gcp.credentials, nil
	}
	jsonKey, err := ioutil.ReadFile(gcp.CredentialsFile)
	if err != nil {
		return nil, err
	}
	credentails, err := google.CredentialsFromJSON(ctx, jsonKey, storage.ScopeFullControl)
	if err != nil {
		return nil, err
	}
	gcp.credentials = credentails
	return credentails, nil
}

func (gcp *GcpConf) getClient(ctx context.Context) (*storage.Client, error) {
	credentails, err := gcp.getCredentials(ctx)
	if err != nil {
		return nil, err
	}
	return storage.NewClient(ctx, option.WithCredentials(credentails))
}

func (gcp *GcpConf) Write(ctx context.Context, key string, pm Perm, writeData func(w io.Writer) error) (path string, err error) {
	client, err := gcp.getClient(ctx)
	if err != nil {
		err = fmt.Errorf("storage.NewClient: %v", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	bucket := gcp.getBucket(pm)

	wc := client.Bucket(bucket).Object(key).NewWriter(ctx)
	if err = writeData(wc); err != nil {
		err = fmt.Errorf("write file error: %s", err.Error())
		return
	}
	if err = wc.Close(); err != nil {
		err = fmt.Errorf("createFile: unable to close bucket %q, file %q: %v", gcp.Bucket, key, err)
		return
	}
	path = wc.Attrs().Name
	return
}

func (gcp *GcpConf) WriteString(ctx context.Context, key string, content string, pm Perm) error {
	client, err := gcp.getClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	bucket := gcp.getBucket(pm)

	wc := client.Bucket(bucket).Object(key).NewWriter(ctx)
	if _, err := wc.Write([]byte(content)); err != nil {
		return fmt.Errorf("createFile: unable to write data to bucket %q, file %q: %v", gcp.Bucket, key, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("createFile: unable to close bucket %q, file %q: %v", gcp.Bucket, key, err)
	}
	return nil
}

func (gcp *GcpConf) RemoveObject(ctx context.Context, key string, pm Perm) error {
	client, err := gcp.getClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bucket := gcp.getBucket(pm)

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	if err := client.Bucket(bucket).Object(key).Delete(ctx); err != nil {
		return fmt.Errorf("delete: unable to delete object bucket %q, file %q: %v", gcp.Bucket, key, err)
	}

	return nil
}

func (gcp *GcpConf) OpenFile(ctx context.Context, key string, pm Perm) (io.Reader, error) {
	client, err := gcp.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	bucket := gcp.getBucket(pm)

	rc, err := client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %v", key, err)
	}
	defer rc.Close()
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %v", err)
	}
	return bytes.NewReader(data), nil
}

func (gcp *GcpConf) GetAttr(ctx context.Context, key string, pm Perm) (*storage.ObjectAttrs, error) {
	client, err := gcp.getClient(ctx)
	if err != nil {
		err = fmt.Errorf("storage.NewClient: %v", err)
		return nil, err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	bucket := gcp.getBucket(pm)
	objectHandle := client.Bucket(bucket).Object(key)
	return objectHandle.Attrs(ctx)
}

func (gcp *GcpConf) GetPublicUrl(ctx context.Context, key string) (myurl string, err error) {
	client, err := gcp.getClient(ctx)
	if err != nil {
		err = fmt.Errorf("storage.NewClient: %v", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	objectHandle := client.Bucket(gcp.PublicBucket).Object(key)
	attrs, err := objectHandle.Attrs(ctx)
	if err != nil {
		return
	}

	u, err := url.Parse(attrs.MediaLink)
	if err != nil {
		return
	}
	rel, err := u.Parse(util.StrAppend("/", gcp.PublicBucket, "/", key))
	if err != nil {
		return
	}
	myurl = rel.String()
	return
}

func (gcp *GcpConf) getBucket(p Perm) string {
	switch p {
	case PermPrivate:
		return gcp.Bucket
	case PermPublic:
		return gcp.PublicBucket
	case PermTmp:
		return gcp.TmpBucket
	}
	return ""
}

func (gcp *GcpConf) SignedURL(key string, contentType string, pm Perm, expDuration time.Duration) (url string, err error) {
	jsonKey, err := ioutil.ReadFile(gcp.CredentialsFile)
	if err != nil {
		return
	}
	conf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		return
	}
	bucket := gcp.getBucket(pm)
	url, err = storage.SignedURL(bucket, key,
		&storage.SignedURLOptions{
			GoogleAccessID: conf.Email,
			Method:         "PUT",
			PrivateKey:     conf.PrivateKey,
			Expires:        time.Now().Add(expDuration),
			ContentType:    contentType,
		})
	return
}

package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/94peter/sterna/util"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
)

type Perm string

const (
	PermPublic  = Perm("public")
	PermPrivate = Perm("private")
	PermTmp     = Perm("tmp")
)

type GcpDI interface {
	NewStorage() (Storage, error)
}

type Storage interface {
	GetAttr(ctx context.Context, key string, pm Perm) (*storage.ObjectAttrs, error)
	RemoveObject(ctx context.Context, key string, pm Perm) error
	GetDownloadUrl(ctx context.Context, key string, p Perm) (myurl string, err error)
	GetPublicUrl(ctx context.Context, object string) (myurl string, err error)
	WriteString(ctx context.Context, key string, content string, pm Perm) error
	Write(ctx context.Context, key string, pm Perm, writeData func(w io.Writer) error) (path string, err error)
	OpenFile(ctx context.Context, key string, pm Perm) (io.Reader, error)
	SignedURL(key string, contentType string, pm Perm, expDuration time.Duration) (url string, err error)
	GetAccessToken() (*oauth2.Token, error)
}

type GcpConf struct {
	CredentialsFile string `yaml:"credentailsFile"`
	CredentailsUrl  string `yaml:"credentailsUrl"`
	Bucket          string
	PublicBucket    string `yaml:"publicBucket"`
	TmpBucket       string `yaml:"tmpBucket"`
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (gcp *GcpConf) NewStorage() (Storage, error) {
	if gcp.CredentailsUrl != "" {
		filePath := fmt.Sprintf("/tmp/%s.json", util.MD5(gcp.CredentailsUrl))
		if !util.FileExists(filePath) {
			err := downloadFile(filePath, gcp.CredentailsUrl)
			if err != nil {
				return nil, err
			}
		}
		gcp.CredentialsFile = filePath
	}
	jsonKey, err := os.ReadFile(gcp.CredentialsFile)
	if err != nil {
		return nil, err
	}
	credentails, err := google.CredentialsFromJSON(context.Background(), jsonKey, storage.ScopeFullControl)
	if err != nil {
		return nil, err
	}

	return &storageImpl{
		GcpConf:     gcp,
		jsonData:    jsonKey,
		credentials: credentails,
	}, nil
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

type storageImpl struct {
	*GcpConf
	jsonData    []byte
	credentials *google.Credentials
}

func (gcp *storageImpl) getClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentials(gcp.credentials))
}

func (gcp *storageImpl) Write(ctx context.Context, key string, pm Perm, writeData func(w io.Writer) error) (path string, err error) {
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

func (gcp *storageImpl) WriteString(ctx context.Context, key string, content string, pm Perm) error {
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

func (gcp *storageImpl) RemoveObject(ctx context.Context, key string, pm Perm) error {
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

func (gcp *storageImpl) OpenFile(ctx context.Context, key string, pm Perm) (io.Reader, error) {
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

func (gcp *storageImpl) GetAttr(ctx context.Context, key string, pm Perm) (*storage.ObjectAttrs, error) {
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

func (gcp *storageImpl) GetPublicUrl(ctx context.Context, key string) (myurl string, err error) {
	return gcp.GetDownloadUrl(ctx, key, PermPublic)
}

func (gcp *storageImpl) GetDownloadUrl(ctx context.Context, key string, p Perm) (myurl string, err error) {
	client, err := gcp.getClient(ctx)
	if err != nil {
		err = fmt.Errorf("storage.NewClient: %v", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	bucket := gcp.getBucket(p)
	objectHandle := client.Bucket(bucket).Object(key)
	attrs, err := objectHandle.Attrs(ctx)
	if err != nil {
		return
	}

	u, err := url.Parse(attrs.MediaLink)
	if err != nil {
		return
	}
	rel, err := u.Parse(util.StrAppend("/", bucket, "/", key))
	if err != nil {
		return
	}
	myurl = rel.String()
	return
}

func (gcp *storageImpl) SignedURL(key string, contentType string, pm Perm, expDuration time.Duration) (url string, err error) {
	conf, err := google.JWTConfigFromJSON(gcp.jsonData)
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

func (gcp *storageImpl) GetAccessToken() (*oauth2.Token, error) {
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal([]byte(gcp.jsonData), &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			"https://www.googleapis.com/auth/devstorage.read_only",
		},
		TokenURL: google.JWTTokenURL,
	}
	token, err := config.TokenSource(context.Background()).Token()
	if err != nil {
		return nil, err
	}
	return token, nil
}

package s3gateway

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	jsoniter "github.com/json-iterator/go"
	"hitminer-file-manager/util"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type object struct {
	FullPath         string `json:"FullPath"`
	Name             string `json:"Name"`
	LastModified     string `json:"LastModified"`
	LastModifiedTime time.Time
	Size             int    `json:"Size"`
	IsDirectory      bool   `json:"IsDirectory"`
	ETag             string `json:"Etag"`
}

func (svr *S3Server) ListObjects(ctx context.Context, prefix string, out io.Writer) error {
	// 如果为根目录,则 prefix = "", 开始不能是 /
	prefix = strings.TrimPrefix(prefix, "/")
	prefix = filepath.Clean(prefix)
	if prefix == "." {
		prefix = ""
	}

	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix = prefix + "/"
	}

	loc := time.Now().Location()
	for object := range svr.listObjects(ctx, prefix, "/") {
		var buffer bytes.Buffer
		if object.IsDirectory {
			buffer.WriteString("drwxr-xr-x\t")
			buffer.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(object.Size))))
			buffer.WriteString(object.LastModifiedTime.In(loc).Format("Jan _2 15:04"))
			buffer.WriteString("\t")
			buffer.WriteString(object.Name)
			buffer.WriteString("\n")
		} else {
			buffer.WriteString("-rwxr-xr--\t")
			buffer.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(object.Size))))
			buffer.WriteString(object.LastModifiedTime.In(loc).Format("Jan 02 15:04"))
			buffer.WriteString("\t")
			buffer.WriteString(object.Name)
			buffer.WriteString("\n")
		}
		_, _ = out.Write(buffer.Bytes())
	}

	svr.mg.Finish()
	return svr.mg.GetError()
}

func (svr *S3Server) listObjects(ctx context.Context, prefix, delimiter string) <-chan object {
	objectCh := make(chan object, 1)

	go func(objectCh chan<- object) {
		defer close(objectCh)
		// Save continuationToken for next request.
		var continuationToken string
		for {
			// Get list of objects a maximum of 1000 per request.
			objects, nextContinuationToken, err := svr.listObjectsReq(ctx, prefix, delimiter, continuationToken)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}

			// If contents are available loop through and send over channel.
			for _, object := range objects {
				object.ETag = util.TrimEtag(object.ETag)
				select {
				// Send object content.
				case objectCh <- *object:
				// If receives done from the caller, return here.
				case <-ctx.Done():
					return
				}
			}

			// If continuation token present, save it for next request.
			if nextContinuationToken != "" {
				continuationToken = nextContinuationToken
			}

			break
		}
	}(objectCh)

	return objectCh
}

func (svr *S3Server) listObjectsReq(ctx context.Context, prefix, delimiter, continuationToken string) ([]*object, string, error) {
	query := url.Values{}
	query.Add("list-type", "2")
	query.Add("prefix", prefix)
	if delimiter != "" {
		query.Add("delimiter", delimiter)
	}
	if continuationToken != "" {
		query.Add("continuation-token", continuationToken)
	}

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/",
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("request error: %d", resp.StatusCode)
	}

	type ret struct {
		NextContinuationToken string    `json:"NextContinuationToken"`
		Objects               []*object `json:"Objects"`
	}
	r := &ret{}

	b, _ := io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(b, r)
	if err != nil {
		return nil, "", err
	}

	for _, object := range r.Objects {
		t, _ := time.Parse("2006-01-02T15:04:05Z07:00", object.LastModified)
		object.LastModifiedTime = t
	}

	return r.Objects, "", nil
}

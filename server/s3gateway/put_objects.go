package s3gateway

import (
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"hitminer-file-manager/util"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	minChunkSize  = 5 * 1024 * 1024
	maxChunkNum   = 9000
	smallFileSize = 5 * 1024 * 1024
)

type part struct {
	Etag             string `json:"Etag"`
	PartNumber       int    `json:"PartNumber"`
	LastModified     string `json:"LastModified"`
	LastModifiedTime time.Time
	Size             int `json:"Size"`
}

func (svr *S3Server) PutObjects(ctx context.Context, filePath, objectName string) error {
	// 如果为根目录,则 prefix = "", 开始不能是 /
	objectName = strings.TrimPrefix(objectName, "/")
	// filepath.Clean会删除最后的/
	if strings.HasSuffix(objectName, "/") {
		objectName = filepath.Clean(objectName) + "/"
	} else {
		objectName = filepath.Clean(objectName)
	}

	fileChan := make(chan string, 1)
	go func(fileChan chan<- string) {
		defer close(fileChan)
		svr.listLocalFile(ctx, filePath, fileChan)
	}(fileChan)

	// filepath: aa/bb/[..]  object: cc/dd   -> cc/dd/..
	// filepath: aa/bb/[..]  object: cc/dd/  -> cc/dd/..
	// filepath: aa/bb[/..]  object: cc/dd   -> cc/dd/..
	// filepath: aa/bb[/..]  object: cc/dd/  -> cc/dd/..
	// filepath: aa/bb       object: cc/dd   -> cc/dd
	// filepath: aa/bb       object: cc/dd/  -> cc/dd/bb
	for fp := range fileChan {
		fullPath := fp
		svr.mg.Add()
		go func() {
			defer svr.mg.Done()
			var remotePath string
			if fullPath == filePath {
				if strings.HasSuffix(objectName, "/") {
					// filepath: aa/bb       object: cc/dd/  -> cc/dd/bb
					remotePath = filepath.ToSlash(filepath.Join(objectName, filepath.Base(fullPath)))
				} else {
					// filepath: aa/bb       object: cc/dd   -> cc/dd
					remotePath = filepath.ToSlash(objectName)
				}
			} else {
				// filepath: aa/bb/[..]  object: cc/dd   -> cc/dd/..
				// filepath: aa/bb/[..]  object: cc/dd/  -> cc/dd/..
				// filepath: aa/bb[/..]  object: cc/dd   -> cc/dd/..
				// filepath: aa/bb[/..]  object: cc/dd/  -> cc/dd/..
				remotePath = filepath.ToSlash(filepath.Join(objectName, fullPath[len(filePath):]))
			}

			f, err := os.Open(fullPath)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}
			defer func() {
				_ = f.Close()
			}()
			stat, err := f.Stat()
			if err != nil {
				svr.mg.AppendError(err)
				return
			}

			if stat.Size() <= smallFileSize {
				err := svr.putObject(ctx, stat.Size(), remotePath, f)
				if err != nil {
					svr.mg.AppendError(err)
					return
				}
			} else {
				err := svr.multiUpload(ctx, stat.Size(), remotePath, "", f)
				if err != nil {
					svr.mg.AppendError(err)
					return
				}
			}
		}()
	}
	svr.mg.Finish()
	return svr.mg.GetError()
}

func (svr *S3Server) listLocalFile(ctx context.Context, dir string, fileChan chan<- string) {
	st, err := os.Stat(dir)
	if err == nil {
		if !st.IsDir() {
			select {
			case fileChan <- dir:
			case <-ctx.Done():
			}
			return
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		svr.mg.AppendError(err)
		return
	}
	for _, ent := range entries {
		if ent.IsDir() {
			svr.listLocalFile(ctx, filepath.Join(dir, ent.Name()), fileChan)
			select {
			case <-ctx.Done():
				return
			default:
			}
		} else {
			select {
			case fileChan <- filepath.Join(dir, ent.Name()):
			case <-ctx.Done():
				return
			}
		}
	}
}

func (svr *S3Server) putObject(ctx context.Context, size int64, objectName string, reader io.Reader) error {
	reqUrl := &url.URL{
		Scheme: "https",
		Host:   svr.host,
		Path:   "/s3/" + objectName,
	}

	bar := util.NewBarReader(reader, size, fmt.Sprintf("upload: %s", objectName))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqUrl.String(), bar)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request error: %d", resp.StatusCode)
	}

	return nil
}

func (svr *S3Server) multiUpload(ctx context.Context, size int64, objectName, uploadId string, reader io.Reader) error {
	reader = util.NewBarReader(reader, size, fmt.Sprintf("upload: %s", objectName))
	parts := make(map[int]*part, 0)
start:
	if uploadId == "" {
		id, err := svr.initMultiUpload(ctx, objectName)
		if err != nil {
			return err
		}
		uploadId = id
	} else {
		mp, err := svr.listMultiUpload(ctx, uploadId, objectName)
		if err != nil {
			uploadId = ""
			goto start
		}
		parts = mp
	}

	chunkSize := util.Max(minChunkSize, size/maxChunkNum)
	for num, offset := 1, int64(0); offset < size; num, offset = num+1, offset+chunkSize {
		partReader := io.LimitReader(reader, chunkSize)
		etag, err := svr.uploadPart(ctx, uploadId, objectName, num, partReader)
		if err != nil {
			return err
		}
		parts[num] = &part{
			Etag:       etag,
			PartNumber: num,
			Size:       int(util.Min(chunkSize, size-offset)),
		}
	}

	err := svr.completeMultiUpload(ctx, uploadId, objectName, parts)
	if err != nil {
		return err
	}

	return nil
}

func (svr *S3Server) initMultiUpload(ctx context.Context, objectName string) (string, error) {
	query := url.Values{}
	query.Add("uploads", "")

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/" + objectName,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error: %d", resp.StatusCode)
	}

	type ret struct {
		Key      string `json:"Key"`
		UploadId string `json:"UploadId"`
	}
	r := &ret{}

	b, _ := io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(b, r)
	if err != nil {
		return "", err
	}

	return r.UploadId, nil
}

func (svr *S3Server) listMultiUpload(ctx context.Context, uploadId, objectName string) (map[int]*part, error) {
	query := url.Values{}
	query.Add("uploadId", uploadId)

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/" + objectName,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: %d", resp.StatusCode)
	}

	type ret struct {
		Part []*part `json:"Part"`
	}
	r := &ret{}

	b, _ := io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}

	parts := make(map[int]*part)
	for i, p := range r.Part {
		t, _ := time.Parse("2006-01-02T15:04:05Z07:00", p.LastModified)
		p.LastModifiedTime = t
		parts[i] = p
	}

	return parts, nil
}

func (svr *S3Server) uploadPart(ctx context.Context, uploadId, objectName string, partNumber int, reader io.Reader) (string, error) {
	query := url.Values{}
	query.Add("uploadId", uploadId)
	query.Add("partNumber", strconv.Itoa(partNumber))

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/" + objectName,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqUrl.String(), reader)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error: %d", resp.StatusCode)
	}

	return resp.Header.Get("Etag"), nil
}

func (svr *S3Server) completeMultiUpload(ctx context.Context, uploadId, objectName string, parts map[int]*part) error {
	query := url.Values{}
	query.Add("uploadId", uploadId)

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/" + objectName,
		RawQuery: query.Encode(),
	}

	type uploadInfo struct {
		Part []struct {
			Etag       string `json:"Etag"`
			PartNumber int    `json:"PartNumber"`
		} `json:"Part"`
	}
	uInfo := &uploadInfo{}
	for _, v := range parts {
		uInfo.Part = append(uInfo.Part, struct {
			Etag       string `json:"Etag"`
			PartNumber int    `json:"PartNumber"`
		}{
			Etag:       util.HTTPEtag(v.Etag),
			PartNumber: v.PartNumber,
		})
	}

	b, err2 := jsoniter.Marshal(uInfo)
	if err2 != nil {
		return err2
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request error: %d", resp.StatusCode)
	}

	return nil
}

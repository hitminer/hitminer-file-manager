package s3gateway

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/hitminer/hitminer-file-manager/server"
	"github.com/hitminer/hitminer-file-manager/util"
	jsoniter "github.com/json-iterator/go"
	md5simd "github.com/minio/md5-simd"
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
	objectName = filepath.ToSlash(objectName)
	objectName = strings.TrimPrefix(objectName, "/")
	// filepath.Clean会删除最后的/
	if strings.HasSuffix(objectName, "/") {
		objectName = filepath.Clean(objectName) + "/"
	} else {
		objectName = filepath.Clean(objectName)
	}
	objectName = filepath.ToSlash(objectName)

	fileChan := make(chan string, 1)
	go func(fileChan chan<- string) {
		defer close(fileChan)
		svr.listLocalFile(ctx, filePath, fileChan)
	}(fileChan)

	listObjects := make(map[string]server.Object)
	for object := range svr.ListObjects(ctx, objectName, "") {
		listObjects[object.FullPath] = object
	}

	haveUploads, err := svr.listMultipartUploads(ctx, objectName)
	if err != nil {
		svr.mg.AppendError(err)
		haveUploads = make(map[string]string)
	}

	md5Sever := md5simd.NewServer()
	defer md5Sever.Close()
	// filepath: aa/bb/[..]  Object: cc/dd   -> cc/dd/..
	// filepath: aa/bb/[..]  Object: cc/dd/  -> cc/dd/..
	// filepath: aa/bb[/..]  Object: cc/dd   -> cc/dd/..
	// filepath: aa/bb[/..]  Object: cc/dd/  -> cc/dd/bb/..
	// filepath: aa/bb       Object: cc/dd   -> cc/dd
	// filepath: aa/bb       Object: cc/dd/  -> cc/dd/bb
	for fp := range fileChan {
		fullPath := fp
		svr.mg.Add()
		go func() {
			defer svr.mg.Done()
			var remotePath string
			if fullPath == filePath {
				if strings.HasSuffix(objectName, "/") {
					// filepath: aa/bb       Object: cc/dd/  -> cc/dd/bb
					remotePath = filepath.ToSlash(filepath.Join(objectName, filepath.Base(fullPath)))
				} else {
					// filepath: aa/bb       Object: cc/dd   -> cc/dd
					remotePath = filepath.ToSlash(objectName)
				}
			} else {
				if strings.HasSuffix(objectName, "/") {
					// filepath: aa/bb[/..]  Object: cc/dd/  -> cc/dd/bb..
					p := 0
					if filepath.Dir(filePath) != "." {
						p = len(filepath.Dir(filePath))
					}
					remotePath = filepath.ToSlash(filepath.Join(objectName, fullPath[p:]))
				} else {
					// filepath: aa/bb/[..]  Object: cc/dd   -> cc/dd/..
					// filepath: aa/bb/[..]  Object: cc/dd/  -> cc/dd/..
					// filepath: aa/bb[/..]  Object: cc/dd   -> cc/dd/..
					remotePath = filepath.ToSlash(filepath.Join(objectName, fullPath[len(filePath):]))
				}
			}

			stat, err := os.Stat(fullPath)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}

			// 判断是否上传过
			if lastUploadInfo, ok := listObjects[remotePath]; ok {
				etag := util.TrimEtag(lastUploadInfo.ETag)
				f, err := os.Open(fullPath)
				if err != nil {
					svr.mg.AppendError(err)
					return
				}
				defer func() {
					_ = f.Close()
				}()
				reader := svr.bar.NewBarReader(f, stat.Size(), fmt.Sprintf("upload check: %s", remotePath))
				index := strings.Index(etag, "-")
				if index == -1 {
					md5Hash := md5Sever.NewHash()
					defer md5Hash.Close()
					_, _ = io.Copy(md5Hash, reader)
					if etag == hex.EncodeToString(md5Hash.Sum(nil)) {
						// 不用上传
						return
					}
				} else {
					etag = etag[:index]
					md5HashSum := md5Sever.NewHash()
					defer md5HashSum.Close()
					chunkSize := util.Max(minChunkSize, stat.Size()/maxChunkNum)
					for offset := int64(0); offset < stat.Size(); offset = offset + chunkSize {
						partReader := io.LimitReader(reader, chunkSize)
						md5Hash := md5Sever.NewHash()
						_, _ = io.Copy(md5Hash, partReader)
						_, _ = md5HashSum.Write(md5Hash.Sum(nil))
						md5Hash.Close()
					}
					if etag == hex.EncodeToString(md5HashSum.Sum(nil)) {
						// 不用上传
						return
					}
				}
			}

			if stat.Size() <= smallFileSize {
				err := svr.putObject(ctx, stat.Size(), remotePath, fullPath)
				if err != nil {
					svr.mg.AppendError(err)
					return
				}
			} else {
				err := svr.multiUpload(ctx, stat.Size(), remotePath, fullPath, haveUploads[remotePath])
				if err != nil {
					svr.mg.AppendError(err)
					return
				}
			}
		}()
	}

	svr.mg.Finish()
	svr.bar.Wait()
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

func (svr *S3Server) putObject(ctx context.Context, size int64, objectName, localPath string) error {
	reqUrl := &url.URL{
		Scheme: "https",
		Host:   svr.host,
		Path:   "/s3/" + objectName,
	}

	reader, err := os.Open(localPath)
	if err != nil {
		return err
	}
	bar := svr.bar.NewBarReader(reader, size, fmt.Sprintf("upload: %s", objectName))
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

func (svr *S3Server) multiUpload(ctx context.Context, size int64, objectName, localPath, uploadId string) error {
	parts := make(map[int]*part, 0)
start:
	if uploadId == "" {
		id, err := svr.initMultiUpload(ctx, objectName)
		if err != nil {
			return err
		}
		uploadId = id
	} else {
		mp, err := svr.listMultipart(ctx, uploadId, objectName)
		if err != nil {
			uploadId = ""
			goto start
		}
		parts = mp
	}
	num, offset := 1, int64(0)
	chunkSize := util.Max(minChunkSize, size/maxChunkNum)
	haveCheck := int64(0)

	if len(parts) != 0 {
		f, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		sumCheckSize := int64(0)
		for _, v := range parts {
			sumCheckSize += int64(v.Size)
		}
		checkReader := svr.bar.NewBarReader(f, sumCheckSize, fmt.Sprintf("upload check: %s", objectName))

		checkParts := make(map[int]*part, 0)
		md5Sever := md5simd.NewServer()
		defer md5Sever.Close()
		for ; offset < size; num, offset = num+1, offset+chunkSize {
			part, ok := parts[num]
			if !ok {
				break
			}
			partReader := io.LimitReader(checkReader, chunkSize)
			md5Hash := md5Sever.NewHash()
			_, _ = io.Copy(md5Hash, partReader)
			etag := hex.EncodeToString(md5Hash.Sum(nil))
			md5Hash.Close()
			if util.TrimEtag(part.Etag) != etag {
				break
			}
			checkParts[num] = part
			haveCheck += int64(part.Size)
		}
		parts = checkParts
	}

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Seek(haveCheck, io.SeekStart)
	if err != nil {
		return err
	}
	reader := svr.bar.NewBarReader(f, size-haveCheck, fmt.Sprintf("upload: %s", objectName))
	for ; offset < size; num, offset = num+1, offset+chunkSize {
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

	err = svr.completeMultiUpload(ctx, uploadId, objectName, parts)
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

func (svr *S3Server) listMultipart(ctx context.Context, uploadId, objectName string) (map[int]*part, error) {
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
	for _, p := range r.Part {
		t, _ := time.Parse("2006-01-02T15:04:05Z07:00", p.LastModified)
		p.LastModifiedTime = t
		p.Etag = util.TrimEtag(p.Etag)
		parts[p.PartNumber] = p
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

func (svr *S3Server) listMultipartUploads(ctx context.Context, prefix string) (map[string]string, error) {
	query := url.Values{}
	query.Add("prefix", prefix)
	query.Add("uploads", "")

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/",
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

	type listUpload struct {
		ListUpload []struct {
			Key       string   `json:"key"`
			UploadIds []string `json:"uploadIds"`
		} `json:"ListUpload"`
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: %d", resp.StatusCode)
	}

	r := &listUpload{}

	b, _ := io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for _, u := range r.ListUpload {
		if len(u.UploadIds) > 0 {
			ret[u.Key] = u.UploadIds[0]
		}
	}

	return ret, nil
}

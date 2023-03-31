package s3gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func (svr *S3Server) RemoveObjects(ctx context.Context, objectName string, recursive bool) error {
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

	if !strings.HasSuffix(objectName, "/") {
		if ok, _ := svr.headObject(ctx, objectName); !ok {
			return fmt.Errorf("not exist file: %s", objectName)
		}
	}

	if strings.HasSuffix(objectName, "/") && recursive == false {
		return fmt.Errorf("removing a directory must be recursive")
	}

	query := url.Values{}
	if recursive {
		query.Add("recursive", "")
	}

	reqUrl := &url.URL{
		Scheme:   "https",
		Host:     svr.host,
		Path:     "/s3/" + objectName,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqUrl.String(), nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("remove failed: %d", resp.StatusCode)
	}

	return nil
}

package s3gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func (svr *S3Server) MakeDirectory(ctx context.Context, objectName string) error {
	objectName = strings.TrimPrefix(objectName, "/")
	objectName = filepath.Clean(objectName)
	if !strings.HasSuffix(objectName, "/") {
		objectName = objectName + "/"
	}

	reqUrl := &url.URL{
		Scheme: "https",
		Host:   svr.host,
		Path:   "/s3/" + objectName,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqUrl.String(), nil)
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

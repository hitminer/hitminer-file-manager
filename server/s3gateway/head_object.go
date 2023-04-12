package s3gateway

import (
	"context"
	"net/http"
	"net/url"
)

func (svr *S3Server) HeadObject(ctx context.Context, objectName string) (bool, int64) {
	reqUrl := &url.URL{
		Scheme: "https",
		Host:   svr.host,
		Path:   "/s3/" + objectName,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqUrl.String(), nil)
	if err != nil {
		return false, 0
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, 0
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode == http.StatusOK, resp.ContentLength
}

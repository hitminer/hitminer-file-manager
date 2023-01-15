package login

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net"
	"net/http"
	"time"
)

type VerifyInfo struct {
	AccessKey string   `json:"accessKey"`
	Endpoints []string `json:"endpoints"`
	Endpoint  string   `json:"endpoint"`
	SecretKey string   `json:"secretKey"`
	Uid       string   `json:"uid"`
}

func Verify(token string) (*VerifyInfo, error) {
	req, err := http.NewRequest("GET", "https://www.hitminer.cn/fileManager/verify", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verify failed")
	}
	ret := &VerifyInfo{}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = jsoniter.Unmarshal(b, ret)
	if err != nil {
		return nil, err
	}
	for _, endpoint := range ret.Endpoints {
		_, err := net.DialTimeout("tcp", endpoint, 1*time.Second)
		if err == nil {
			ret.Endpoint = endpoint
			break
		}
	}

	if ret.Endpoint == "" {
		return nil, fmt.Errorf("No network available")
	}

	return ret, nil
}

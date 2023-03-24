package login

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
)

func Login(host, username, password string) (string, error) {
	type Info struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"rememberMe"`
	}
	info := &Info{
		Username:   username,
		Password:   password,
		RememberMe: true,
	}
	b, err := jsoniter.Marshal(info)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("https://"+host+"/bizapi/bizuser/login", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login error")
	}
	type Ret struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Token string `json:"token"`
	}
	ret := &Ret{}
	b, _ = io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(b, ret)
	if err != nil {
		return "", err
	}
	if ret.Code != http.StatusOK {
		return "", fmt.Errorf("login failed")
	}
	return ret.Token, nil
}

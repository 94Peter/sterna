package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/94peter/sterna/util"
)

type TransmitSecurity interface {
	GetAuthUrl(redirect string) string
	GetAccessToken(code, redirect string) (string, error)
	GetUserInfo(accessToken string) (string, error)
}

type TransmitSecurityConf struct {
	Host     string
	ClientId string `yaml:"clientId"`
	Secret   string `yaml:"clientSecret"`
}

func (c *TransmitSecurityConf) GetAuthUrl(redirect string) string {
	params := url.Values{
		"client_id":     {c.ClientId},
		"redirect_uri":  {redirect},
		"scope":         {"openid email"},
		"response_type": {"code"},
	}
	return fmt.Sprintf("https://%s/authorize?%s", c.Host, params.Encode())
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (c *TransmitSecurityConf) GetAccessToken(code, redirect string) (string, error) {
	params := url.Values{
		"acr_values":    {"ts.bindid.iac.email"},
		"redirect_uri":  {redirect},
		"client_id":     {c.ClientId},
		"client_secret": {c.Secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(fmt.Sprintf("https://%s/token", c.Host), params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tokenResp := tokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func (c *TransmitSecurityConf) GetUserInfo(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/userinfo", c.Host), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", util.StrAppend("Bearer ", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
	return "", nil
}

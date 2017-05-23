package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type User struct {
	ID     string        `json:"id"`
	Email  string        `json:"email"`
	Token  *oauth2.Token `json:"token"`
	client *http.Client  `json:"-"`
}

type Me struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type ResumeList struct {
	Resumes []*Resume `json:"items"`
}

type Resume struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (u *User) SetClient(client *http.Client) {
	u.client = client
}

func (u *User) CanRefresh() bool {
	return time.Now().After(u.Token.Expiry)
}

func (u *User) GetResumeList() ([]*Resume, error) {
	var resultList []*Resume
	resp, err := u.client.Get("https://api.hh.ru/resumes/mine")
	if err != nil {
		return resultList, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resultList, err
	}
	var rl ResumeList
	if err := json.Unmarshal(body, &rl); err != nil {
		return resultList, err
	}
	return rl.Resumes, nil
}

func (u *User) PublishResume(r *Resume) (string, error) {
	var result string
	url := fmt.Sprintf("https://api.hh.ru/resumes/%s/publish", r.ID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.Token.AccessToken))
	resp, err := u.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	if code := resp.StatusCode; code < 200 || code > 299 {
		return result, fmt.Errorf("Incorrect status code (%s)", resp.Status)
	}
	return string(body), nil
}

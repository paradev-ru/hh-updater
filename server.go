package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/leominov/hh"
)

var (
	UsersBucket = []byte("usersv1")
	UsersKey    = []byte("list")
)

type Server struct {
	c         *Config
	userList  map[string]*User
	oAuthConf *oauth2.Config
	db        *bolt.DB
}

func NewServer(config *Config) *Server {
	return &Server{
		c:        config,
		userList: map[string]*User{},
		oAuthConf: &oauth2.Config{
			Endpoint:     hh.Endpoint,
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
		},
	}
}

func (s *Server) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	url := s.oAuthConf.AuthCodeURL(s.c.StateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	q := r.URL.Query()

	if q.Get("state") != s.c.StateString {
		return nil, errors.New("Invalid oAuth2 state")
	}

	token, err := s.oAuthConf.Exchange(oauth2.NoContext, q.Get("code"))
	if err != nil {
		return nil, err
	}

	// https://github.com/hhru/api/blob/master/docs/me.md#Получение-информации-о-текущем-пользователе
	client := s.oAuthConf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://api.hh.ru/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var me Me
	if err := json.Unmarshal(body, &me); err != nil {
		return nil, err
	}

	if _, ok := s.userList[me.ID]; !ok {
		logrus.Infof("Add user: %s", me.Email)
	}

	u := &User{
		ID:    me.ID,
		Email: me.Email,
		Token: token,
	}
	u.SetClient(s.oAuthConf.Client(oauth2.NoContext, u.Token))

	return u, nil
}

func (s *Server) HandleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := s.handleUser(w, r)
	if err != nil {
		logrus.Error(err)
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	}
	s.userList[user.ID] = user
	http.Redirect(w, r, "/logged.html", http.StatusFound)
}

func (s *Server) Daemon() {
	for {
		for _, user := range s.userList {
			logrus.Debugf("Getting resume for %s...", user.Email)
			resumeList, err := user.GetResumeList()
			if err != nil {
				logrus.Errorf("Error getting resume list for %s: %v", user.Email, err)
				continue
			}
			time.Sleep(s.c.RequestSleep)
			for _, resume := range resumeList {
				logrus.Debugf("Publishing resume '%s'...", resume.Title)
				_, err := user.PublishResume(resume)
				if err != nil {
					logrus.Errorf("Error publishing resume '%s': %v", resume.Title, err)
					time.Sleep(s.c.RequestSleep)
					continue
				}
				logrus.Debugf("Resume '%s' updated", resume.Title)
				time.Sleep(s.c.RequestSleep)
			}
		}
		time.Sleep(s.c.LoopSleep)
	}
}

func (s *Server) Init() error {
	db, err := bolt.Open(s.c.DatabasePath, 0600, nil)
	if err != nil {
		return err
	}

	s.db = db

	return s.db.Update(func(tx *bolt.Tx) error {
		// Always create Users bucket.
		if _, err := tx.CreateBucketIfNotExists(UsersBucket); err != nil {
			return err
		}
		return nil
	})
}

func (s *Server) ResoreUserList() error {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UsersBucket)
		v := b.Get(UsersKey)
		if len(v) == 0 {
			logrus.Warn("No entries in database")
			return nil
		}
		if err := json.Unmarshal(v, &s.userList); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for id, u := range s.userList {
		u.client = s.oAuthConf.Client(oauth2.NoContext, u.Token)
		s.userList[id] = u
	}
	return nil
}

func (s *Server) SaveUserList() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(UsersBucket)
		encoded, err := json.Marshal(s.userList)
		if err != nil {
			return err
		}
		return b.Put(UsersKey, encoded)
	})
}

func (s *Server) Stop() error {
	return s.SaveUserList()
}

func (s *Server) Start() error {
	if err := s.ResoreUserList(); err != nil {
		return err
	}

	http.HandleFunc("/authorize", s.HandleAuthorize)
	http.HandleFunc("/callback", s.HandleCallback)

	http.Handle("/", http.FileServer(http.Dir("./public")))

	go s.Daemon()

	logrus.Infof("Started running on http://%s", s.c.ListenAddress)
	return http.ListenAndServe(s.c.ListenAddress, nil)
}

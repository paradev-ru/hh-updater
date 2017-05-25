package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/leominov/hh"
	"github.com/paradev-ru/hh-updater/hhclient"
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

type User struct {
	ID    string        `json:"id"`
	Email string        `json:"email"`
	Token *oauth2.Token `json:"token"`
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
	client := hhclient.NewClient(token)
	me, err := client.Me.GetMe()
	if err != nil {
		return nil, err
	}
	u := &User{
		ID:    me.ID,
		Email: me.Email,
		Token: token,
	}
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

func (s *Server) ResumePublishDaemon() {
	for {
		for i, user := range s.userList {
			logrus.Debugf("Getting information of user: %s", user.Email)
			tokenSource := s.oAuthConf.TokenSource(oauth2.NoContext, user.Token)
			newToken, err := tokenSource.Token()
			if err != nil {
				logrus.Errorf("Error getting token for user %s: %v", user.Email, err)
				continue
			}
			if user.Token.AccessToken != newToken.AccessToken {
				logrus.Infof("Updating token for user %s", user.Email)
				user.Token = newToken
				s.userList[i] = user
				logrus.Infof("New expiry date for user %s token: %s", user.Email, user.Token.Expiry.String())
			}
			client := hhclient.NewClient(user.Token)
			if _, err := client.Me.GetMe(); err != nil {
				logrus.Errorf("Error getting information of user %s: %v", user.Email, err)
				continue
			}
			logrus.Debugf("Getting resumes for user: %s", user.Email)
			resumeList, err := client.Resume.ResumeMine()
			if err != nil {
				logrus.Errorf("Error getting resume for user %s: %v", user.Email, err)
				continue
			}
			for _, r := range resumeList {
				logrus.Debugf("Requesting resume status: '%s'", r.Title)
				status, err := client.Resume.ResumesStatus(r)
				if err != nil {
					logrus.Errorf("Error getting resume status '%s': %v", r.Title, err)
					continue
				}
				if !status.CanPublishOrUpdate {
					logrus.Debugf("Skipping publish resume: '%s'", r.Title)
					continue
				}
				logrus.Debugf("Publishing resume: '%s'", r.Title)
				if err := client.Resume.ResumesPublish(r); err != nil {
					logrus.Errorf("Error publishing resume '%s': %v", r.Title, err)
					continue
				}
				logrus.Infof("Resume updated: '%s'", r.Title)
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
	return s.db.View(func(tx *bolt.Tx) error {
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

	go s.ResumePublishDaemon()

	logrus.Infof("Started running on %s", s.c.ListenAddress)
	return http.ListenAndServe(s.c.ListenAddress, nil)
}

package main

import "fmt"

type ResumeStatus struct {
	Blocked            bool   `json:"blocked"`
	Finished           bool   `json:"finished"`
	CanPublishOrUpdate bool   `json:"can_publish_or_update"`
	PublishURL         string `json:"publish_url"`
}

func (r *ResumeStatus) IsPublishAllowed() bool {
	if r.Blocked == true {
		return false
	}
	if r.Finished == false {
		return false
	}
	if r.CanPublishOrUpdate == false {
		return false
	}
	return true
}

func (r *ResumeStatus) String() string {
	return fmt.Sprintf(
		"Blocked:%v, Finished:%v, CanPublishOrUpdate:%v",
		r.Blocked,
		r.Finished,
		r.CanPublishOrUpdate,
	)
}

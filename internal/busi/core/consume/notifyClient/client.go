package notifyClient

import (
	"errors"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

var (
	reqClient *req.Client
)

func init() {
	reqClient = req.C().SetTimeout(5 * time.Second)
}

type ErrResponse struct {
	RequestID string `json:"request_id"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

type Topic struct {
	Topic string `form:"topic" json:"topic" binding:"required" example:"messages/vm_messages..."`
}

type TipsetState struct {
	Topic         string `form:"topic" json:"topic" binding:"required"`
	Tipset        uint64 `form:"tipset" json:"tipset" desc:"tipset.Height()"`
	Version       uint16 `form:"version" json:"version"`
	State         uint16 `form:"state" json:"state" desc:"1 - task successful, 2 - task failed"`
	NotFoundState uint8  `form:"not_found_state" json:"not_found_state" desc:"1 - tipset not found, can't find tipset is a special case of failed tasks, when state equal 2, this field takes effect"`
	Description   string `form:"description" json:"description"`
}

// TopicSignIn register topic
func TopicSignIn(host string, topic string) error {
	var result ErrResponse

	resp, err := reqClient.R().SetBody(&Topic{Topic: topic}).SetSuccessResult(&result).Post(fmt.Sprintf("%s/api/v1/topic", host))
	if err != nil {
		return err
	}
	if !resp.IsSuccessState() {
		return errors.New(fmt.Sprintf("bad response status: %v", resp.Status))
	}

	return nil
}

// ReportTipsetState report chain notify server task state
func ReportTipsetState(host string, force bool, topic string, height int64, version, state, notFoundState int, desc string) error {
	var result ErrResponse

	tipsetState := TipsetState{
		Topic:         topic,
		Tipset:        uint64(height),
		Version:       uint16(version),
		State:         uint16(state),
		NotFoundState: uint8(notFoundState),
		Description:   desc,
	}
	resp, err := reqClient.R().SetBody(&tipsetState).SetSuccessResult(&result).Post(fmt.Sprintf("%s/api/v1/task_state?force=%v", host, force))
	if err != nil {
		return err
	}
	if !resp.IsSuccessState() {
		return errors.New(fmt.Sprintf("bad response status: %v", resp.Status))
	}
	return nil
}

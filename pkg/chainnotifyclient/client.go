package chainnotifyclient

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// TopicSignIn register topic
func TopicSignIn(host string, topic string) error {
	params := map[string]string{
		"topic": topic,
	}
	resp := req.C().Post(fmt.Sprintf("%s/api/v1/topic", host)).SetBodyJsonMarshal(params).Do()
	return resp.Err
}

// ReportTipsetState report chain notify server task state
func ReportTipsetState(host string, topic string, height, version, state, notFoundState int, desc string) error {
	params := map[string]interface{}{
		"topic":           topic,
		"tipset":          height,
		"version":         version,
		"state":           state,
		"not_found_state": notFoundState,
		"description":     desc,
	}
	resp := req.C().Post(fmt.Sprintf("%s/api/v1/task_state", host)).SetBodyJsonMarshal(params).Do()
	return resp.Err
}

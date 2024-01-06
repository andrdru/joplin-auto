package webClipperClient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	WebClipper struct {
		client *http.Client

		url      string
		apiToken string
	}

	errorResp struct {
		Error string `json:"error"`
	}

	Note struct {
		ID       string `json:"id"`
		ParentID string `json:"parent_id"`
		Title    string `json:"title"`
		Body     string `json:"body"`
	}

	List struct {
		Items   []Note `json:"items"`
		HasMore bool   `json:"has_more"`
	}
)

var (
	ErrCodeServerError = errors.New("server error")
	ErrCodeLogicError  = errors.New("logic error")
)

func NewWebClipper(serviceUrl string, apiToken string) *WebClipper {
	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			DialContext: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	return &WebClipper{
		client:   client,
		url:      serviceUrl,
		apiToken: apiToken,
	}
}

func (w *WebClipper) List(ctx context.Context, page int) (list List, err error) {
	q := url.Values{}
	q.Add("token", w.apiToken)
	q.Add("fields", "id,parent_id")
	q.Add("page", fmt.Sprintf("%d", page))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url+"/notes?"+q.Encode(), nil)
	if err != nil {
		return List{}, fmt.Errorf("new request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return List{}, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := readResp(resp)
	if err != nil {
		return List{}, fmt.Errorf("read response: %w", err)
	}

	err = json.Unmarshal(data, &list)
	if err != nil {
		return List{}, fmt.Errorf("unmarshal: %w", err)
	}

	return list, nil
}

func (w *WebClipper) Get(ctx context.Context, id string) (note Note, err error) {
	q := url.Values{}
	q.Add("token", w.apiToken)
	q.Add("fields", "id,parent_id,body,title")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url+"/notes/"+id+"?"+q.Encode(), nil)
	if err != nil {
		return Note{}, fmt.Errorf("new request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return Note{}, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := readResp(resp)
	if err != nil {
		return Note{}, fmt.Errorf("read response: %w", err)
	}

	err = json.Unmarshal(data, &note)
	if err != nil {
		return Note{}, fmt.Errorf("unmarshal: %w", err)
	}

	return note, nil
}

func (w *WebClipper) Put(ctx context.Context, id string, body string) (note Note, err error) {
	q := url.Values{}
	q.Add("token", w.apiToken)
	q.Add("body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, w.url+"/notes/"+id, strings.NewReader(q.Encode()))
	if err != nil {
		return Note{}, fmt.Errorf("new request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return Note{}, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := readResp(resp)
	if err != nil {
		return Note{}, fmt.Errorf("read response: %w", err)
	}

	err = json.Unmarshal(data, &note)
	if err != nil {
		return Note{}, fmt.Errorf("unmarshal: %w", err)
	}

	return note, nil
}

func readResp(resp *http.Response) (data []byte, err error) {
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("%d: %w", resp.StatusCode, ErrCodeServerError)
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResp
		if errUnmarshal := json.Unmarshal(data, &errResp); errUnmarshal != nil {
			return nil, fmt.Errorf("unmarshal: %s: %w", errUnmarshal, ErrCodeLogicError)
		}

		return nil, fmt.Errorf("%s: %w", errResp.Error, ErrCodeLogicError)
	}

	return data, nil
}

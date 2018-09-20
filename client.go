package aliyun_mns

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	endpoint        string
	accessKeyId     string
	accessKeySecret string
	queues          []*Queue
	doneQueues      map[string]struct{}
}

func New(endpoint, accessKeyId, accessKeySecret string) *Client {
	client := new(Client)
	client.endpoint = endpoint
	client.accessKeyId = accessKeyId
	client.accessKeySecret = accessKeySecret
	client.queues = make([]*Queue, 0)
	client.doneQueues = make(map[string]struct{})
	return client
}

// https://help.aliyun.com/document_detail/27485.html?spm=a2c4g.11186623.6.694.15043ca8X9WLlR
func (c *Client) finalizeHeader(request *http.Request, body []byte) {
	request.Header.Set(contentType, textXml)
	request.Header.Set(date, time.Now().UTC().Format(http.TimeFormat))
	request.Header.Set(xMnsVersion, "2015-06-06")
	request.Header.Set(contentLength, strconv.FormatInt(int64(len(body)), 10))
	request.Header.Set(host, request.Host)
	request.Header.Set(contentMd5, Base64Md5(string(body)))
	toSign := []byte(
		request.Method + "\n" +
			request.Header.Get(contentMd5) + "\n" +
			textXml + "\n" +
			request.Header.Get(date) + "\n" +
			c.canonicalizedMNSHeaders(request.Header) + "\n" +
			c.canonicalizedResource(request),
	)
	auth := fmt.Sprintf("MNS %s:%s", c.accessKeyId, c.signature(toSign))
	request.Header.Set(authorization, auth)
}

func (c *Client) signature(toSign []byte) string {
	h := hmac.New(sha1.New, []byte(c.accessKeySecret))
	h.Write(toSign)
	sign := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sign)
}

func (c *Client) canonicalizedMNSHeaders(header http.Header) string {
	var mnsHeaders []string
	for k, _ := range header {
		if l := strings.ToLower(k); strings.HasPrefix(l, "x-mns-") {
			mnsHeaders = append(mnsHeaders, l)
		}
	}
	sort.Strings(mnsHeaders)
	var pairs []string
	for _, mnsHeader := range mnsHeaders {
		pairs = append(pairs, mnsHeader+":"+header.Get(mnsHeader))
	}
	return strings.Join(pairs, "\n")
}

func (c *Client) canonicalizedResource(request *http.Request) string {
	if request.URL.RawQuery == "" {
		return request.URL.Path
	}
	return request.URL.Path + "?" + request.URL.RawQuery
}


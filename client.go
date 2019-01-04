package alimns

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Client 存储了阿里云的相关信息
type Client struct {
	queuePrefix     string
	endpoint        string
	accessKeyID     string
	accessKeySecret string
}

// NewClient 返回Client的实例
func NewClient(endpoint, accessKeyID, accessKeySecret string) Client {
	return Client{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}
}

// SetQueuePrefix sets the query param for ListQueue.
func (c *Client) SetQueuePrefix(prefix string) {
	c.queuePrefix = prefix
}

// Base64Md5 md5值用base64编码
func Base64Md5(s string) string {
	hash := md5.New()
	io.WriteString(hash, s)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// https://help.aliyun.com/document_detail/27485.html?spm=a2c4g.11186623.6.694.15043ca8X9WLlR
func (c *Client) finalizeHeader(request *http.Request, body []byte) {
	request.Header.Set(contentType, textXML)
	request.Header.Set(date, time.Now().UTC().Format(http.TimeFormat))
	request.Header.Set(xMnsVersion, "2015-06-06")
	request.Header.Set(contentLength, strconv.FormatInt(int64(len(body)), 10))
	request.Header.Set(host, request.Host)
	request.Header.Set(contentMd5, Base64Md5(string(body)))
	toSign := []byte(
		request.Method + "\n" +
			request.Header.Get(contentMd5) + "\n" +
			textXML + "\n" +
			request.Header.Get(date) + "\n" +
			c.canonicalizedMNSHeaders(request.Header) + "\n" +
			c.canonicalizedResource(request),
	)
	auth := fmt.Sprintf("MNS %s:%s", c.accessKeyID, c.signature(toSign))
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
	for k := range header {
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

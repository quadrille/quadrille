package client

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type request struct {
	url              string
	headers          map[string]string
	cookiesStr       string
	returnCookies    bool
	timeout          int
	httpMethod       string
	payload          string
	contentType      ContentType
	queryParams      map[string]string
	preventRedirects bool
}

//ContentType ...
type ContentType string

const (
	JSON           ContentType = "application/json"
	FormURLEncoded ContentType = "application/x-www-form-urlencoded"
)

//Get ...
func Get(url string) request {
	return request{url: url}
}

//Post ...
func Post(url string) request {
	return request{url: url, httpMethod: "POST"}
}

//Put ...
func Put(url string) request {
	return request{url: url, httpMethod: "PUT"}
}

//Head ...
func Head(url string) request {
	return request{url: url, httpMethod: "HEAD"}
}

//Delete ...
func Delete(url string) request {
	return request{url: url, httpMethod: "DELETE"}
}

func (r request) PreserveCookies() request {
	r.returnCookies = true
	return r
}

func getCookies(str string) string {
	cookieStr := ""
	var re = regexp.MustCompile(`(?m)(.*?)=(.*?);`)

	for i, match := range re.FindAllString(str, -1) {
		fmt.Println(match, "found at index", i)
		if !strings.Contains(match, "path") && !strings.Contains(match, "expires") {
			cookieStr += match
		}

	}
	return cookieStr
}

func (r request) SetHeaders(headers map[string]string) request {
	r.headers = headers
	return r
}

func (r request) SetCookies(cookies string) request {
	r.cookiesStr = cookies
	return r
}

func (r request) SetPayload(payload string) request {
	r.payload = payload
	return r
}

func (r request) SetTimeout(timeout int) request {
	r.timeout = timeout
	return r
}

func (r request) SetContentType(contentType ContentType) request {
	r.contentType = contentType
	return r
}

func (r request) SetQueryParams(queryParams map[string]string) request {
	r.queryParams = queryParams
	return r
}

func (r request) PreventRedirects() request {
	r.preventRedirects = true
	return r
}

func (r request) Do() (responseBody string, cookie string, err error) {
	timeout := time.Second * 10
	if r.timeout != 0 {
		timeout = time.Second * time.Duration(r.timeout)
	}
	client := &http.Client{Timeout: timeout}

	if r.preventRedirects {
		client = &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}, Timeout: timeout}
	}
	httpMethod := "GET"
	if r.httpMethod != "" {
		httpMethod = r.httpMethod
	}
	var body io.Reader
	if r.payload != "" {
		body = strings.NewReader(r.payload)
	}
	req, err := http.NewRequest(httpMethod, r.url, body)
	if r.contentType != "" {
		req.Header.Add("Content-Type", ((string)(r.contentType)))
	}
	if r.cookiesStr != "" {
		req.Header.Add("Cookie", r.cookiesStr)
	}

	if len(r.headers) > 0 {
		for k, v := range r.headers {
			req.Header.Add(k, v)
		}
	}

	if r.queryParams != nil {
		q := req.URL.Query()
		for k, v := range r.queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", "", err
	} else if resp.StatusCode > 400 {
		return "", "", errors.New("Response " + strconv.Itoa(resp.StatusCode))
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	if err != nil {
		fmt.Println(err)
	}
	cookiesStr := ""
	if r.returnCookies {
		for k, v := range resp.Header {
			key := strings.ToLower(k)
			if key == "set-cookie" {
				cookiesStr += string(strings.Join(v, " "))
			}
		}
		cookiesStr = strings.Replace(cookiesStr, "secure", "", -1)
		cookiesStr = strings.Replace(cookiesStr, "HttpOnly", "", -1)
	}
	return bodyString, cookiesStr, err
}

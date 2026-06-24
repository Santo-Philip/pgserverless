package proxy

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PostgRESTProxy struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func NewPostgRESTProxy(baseURL string, timeout time.Duration) *PostgRESTProxy {
	return &PostgRESTProxy{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		timeout: timeout,
	}
}

func (p *PostgRESTProxy) Forward(c *fiber.Ctx, appSlug string) error {
	targetPath := strings.TrimPrefix(c.Path(), "/v1/"+appSlug)
	if targetPath == "" {
		targetPath = "/"
	}

	queryString := string(c.Request().URI().QueryString())
	if queryString != "" {
		targetPath += "?" + queryString
	}

	targetURL := p.baseURL + targetPath

	bodyBytes := c.Body()
	var reqBody io.Reader
	if len(bodyBytes) > 0 {
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(c.Method(), targetURL, reqBody)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to create request")
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		skey := string(key)
		if skey == "Host" || skey == "X-API-Key" {
			return
		}
		req.Header.Set(skey, string(value))
	})

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Header.Set("X-Request-ID", c.Locals("request_id").(string))

	if appSlug != "" {
		req.Header.Set("Accept-Profile", appSlug)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "upstream error: "+err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to read response")
	}

	for key, values := range resp.Header {
		for _, v := range values {
			skey := http.CanonicalHeaderKey(key)
			if skey == "Content-Length" {
				continue
			}
			c.Set(key, v)
		}
	}

	return c.Status(resp.StatusCode).Send(respBody)
}

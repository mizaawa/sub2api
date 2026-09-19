package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ChannelMonitorTimestampHeader = "X-Sub2API-Monitor-Timestamp"
	ChannelMonitorSignatureHeader = "X-Sub2API-Monitor-Signature"

	channelMonitorAttestationLabel = "sub2api/channel-monitor-attestation/v1"
	channelMonitorAttestationSkew  = 90 * time.Second
)

var ErrChannelMonitorAttestationUnavailable = errors.New("channel monitor attestation is unavailable")

// ChannelMonitorAttestor proves that a request using a managed monitor key was
// created by this deployment's monitor runner. The raw API key alone is not
// sufficient to use a managed credential.
type ChannelMonitorAttestor struct {
	key []byte
}

// NewChannelMonitorAttestor derives a domain-separated HMAC key from the
// deployment encryption key. The configured value is expected to be the same
// 32-byte hex key used by SecretEncryptor.
func NewChannelMonitorAttestor(encryptionKeyHex string) (*ChannelMonitorAttestor, error) {
	master, err := hex.DecodeString(strings.TrimSpace(encryptionKeyHex))
	if err != nil || len(master) != 32 {
		return nil, ErrChannelMonitorAttestationUnavailable
	}
	mac := hmac.New(sha256.New, master)
	_, _ = mac.Write([]byte(channelMonitorAttestationLabel))
	return &ChannelMonitorAttestor{key: mac.Sum(nil)}, nil
}

// SignRequest adds a short-lived request signature. body must be the exact
// bytes that will be sent by req.
func (a *ChannelMonitorAttestor) SignRequest(req *http.Request, apiKey string, body []byte) error {
	return a.signRequestAt(req, apiKey, body, time.Now())
}

func (a *ChannelMonitorAttestor) signRequestAt(req *http.Request, apiKey string, body []byte, now time.Time) error {
	if a == nil || len(a.key) == 0 || req == nil || req.URL == nil || strings.TrimSpace(apiKey) == "" {
		return ErrChannelMonitorAttestationUnavailable
	}
	timestamp := strconv.FormatInt(now.Unix(), 10)
	signature := a.signature(timestamp, req.Method, requestAttestationPath(req), body, apiKey)
	req.Header.Set(ChannelMonitorTimestampHeader, timestamp)
	req.Header.Set(ChannelMonitorSignatureHeader, signature)
	return nil
}

// ValidateRequest verifies and then restores the request body for downstream
// handlers. Global gateway body limits are installed before authentication, so
// this read remains bounded by the normal ingress limit.
func (a *ChannelMonitorAttestor) ValidateRequest(req *http.Request, apiKey string) bool {
	if a == nil || len(a.key) == 0 || req == nil || req.URL == nil || strings.TrimSpace(apiKey) == "" {
		return false
	}
	timestamp := strings.TrimSpace(req.Header.Get(ChannelMonitorTimestampHeader))
	provided := strings.TrimSpace(req.Header.Get(ChannelMonitorSignatureHeader))
	unixSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || provided == "" {
		return false
	}
	delta := time.Since(time.Unix(unixSeconds, 0))
	if delta < -channelMonitorAttestationSkew || delta > channelMonitorAttestationSkew {
		return false
	}

	body, err := readAndRestoreRequestBody(req)
	if err != nil {
		return false
	}
	expected := a.signature(timestamp, req.Method, requestAttestationPath(req), body, apiKey)
	providedBytes, err := base64.RawURLEncoding.DecodeString(provided)
	if err != nil {
		return false
	}
	expectedBytes, err := base64.RawURLEncoding.DecodeString(expected)
	return err == nil && hmac.Equal(providedBytes, expectedBytes)
}

func (a *ChannelMonitorAttestor) signature(timestamp, method, path string, body []byte, apiKey string) string {
	bodyHash := sha256.Sum256(body)
	keyHash := sha256.Sum256([]byte(apiKey))
	canonical := strings.Join([]string{
		timestamp,
		strings.ToUpper(strings.TrimSpace(method)),
		path,
		hex.EncodeToString(bodyHash[:]),
		hex.EncodeToString(keyHash[:]),
	}, "\n")
	mac := hmac.New(sha256.New, a.key)
	_, _ = mac.Write([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func requestAttestationPath(req *http.Request) string {
	path := req.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	return path
}

func readAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read attested request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

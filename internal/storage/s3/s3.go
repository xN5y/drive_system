package s3

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"simple-drive/internal/models"
)

type S3Storage struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	region    string
	client    *http.Client
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket, region string) (*S3Storage, error) {
	s3 := &S3Storage{
		endpoint:  strings.TrimSuffix(endpoint, "/"),
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		region:    region,
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	if err := s3.createBucketIfNotExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return s3, nil
}

func (s *S3Storage) Save(blob *models.Blob) (string, error) {
	var url string
	if strings.Contains(s.endpoint, s.bucket) {
		url = fmt.Sprintf("%s/%s", s.endpoint, blob.ID)
	} else {
		url = fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, blob.ID)
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(blob.Data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()
	s.signRequest(req, blob.Data, now)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("S3 upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return blob.ID, nil
}

func (s *S3Storage) Retrieve(id string) (*models.Blob, error) {
	var url string
	if strings.Contains(s.endpoint, s.bucket) {
		url = fmt.Sprintf("%s/%s", s.endpoint, id)
	} else {
		url = fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, id)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()
	s.signRequest(req, nil, now)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("S3 retrieval failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &models.Blob{
		ID:   id,
		Data: data,
	}, nil
}

func (s *S3Storage) createBucketIfNotExists() error {
	if strings.Contains(s.endpoint, s.bucket) {
		return nil
	}
	url := fmt.Sprintf("%s/%s", s.endpoint, s.bucket)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	s.signRequest(req, nil, now)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create bucket (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *S3Storage) signRequest(req *http.Request, payload []byte, now time.Time) {
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	var payloadHash string
	if payload != nil {
		hash := sha256.Sum256(payload)
		payloadHash = hex.EncodeToString(hash[:])
	} else {
		hash := sha256.Sum256([]byte{})
		payloadHash = hex.EncodeToString(hash[:])
	}

	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	canonicalRequest := s.buildCanonicalRequest(req, payloadHash)
	stringToSign := s.buildStringToSign(canonicalRequest, amzDate, dateStamp)
	signature := s.calculateSignature(stringToSign, dateStamp)

	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", s.accessKey, dateStamp, s.region)
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s",
		credential, s.getSignedHeaders(req), signature)

	req.Header.Set("Authorization", authHeader)
}

func (s *S3Storage) buildCanonicalRequest(req *http.Request, payloadHash string) string {
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalQueryString := req.URL.RawQuery

	var headerNames []string
	headerMap := make(map[string]string)
	for name, values := range req.Header {
		lowerName := strings.ToLower(name)
		headerNames = append(headerNames, lowerName)
		headerMap[lowerName] = strings.Join(values, ",")
	}
	sort.Strings(headerNames)

	var canonicalHeaders strings.Builder
	for _, name := range headerNames {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(headerMap[name]))
		canonicalHeaders.WriteString("\n")
	}

	signedHeaders := strings.Join(headerNames, ";")

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash)
}

func (s *S3Storage) buildStringToSign(canonicalRequest, amzDate, dateStamp string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHash := hex.EncodeToString(hash[:])

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.region)

	return fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		credentialScope,
		canonicalRequestHash)
}

func (s *S3Storage) calculateSignature(stringToSign, dateStamp string) string {
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	signature := hmacSHA256(kSigning, []byte(stringToSign))

	return hex.EncodeToString(signature)
}

func (s *S3Storage) getSignedHeaders(req *http.Request) string {
	var headerNames []string
	for name := range req.Header {
		headerNames = append(headerNames, strings.ToLower(name))
	}
	sort.Strings(headerNames)
	return strings.Join(headerNames, ";")
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

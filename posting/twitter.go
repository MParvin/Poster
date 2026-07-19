package posting

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TwitterCredentials holds the necessary credentials for posting to Twitter.
type TwitterCredentials struct {
	APIKey            string
	APISecretKey      string
	AccessToken       string
	AccessTokenSecret string
}

// PostToTwitter sends a tweet using the Twitter API v2.
func PostToTwitter(messageContent string, creds TwitterCredentials) error {
	if creds.APIKey == "" || creds.APISecretKey == "" || creds.AccessToken == "" || creds.AccessTokenSecret == "" {
		return fmt.Errorf("twitter credentials are not fully configured")
	}

	endpoint := "https://api.twitter.com/2/tweets"
	payload := map[string]string{"text": messageContent}
	bodyBytes, err := marshalJSON(payload)
	if err != nil {
		return err
	}

	authHeader, err := buildOAuth1Header("POST", endpoint, nil, creds)
	if err != nil {
		return fmt.Errorf("build twitter auth header: %w", err)
	}

	log.Printf("[TWITTER] Posting tweet (%d chars, key %s)", len(messageContent), truncateToken(creds.APIKey))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("create twitter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("twitter request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read twitter response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twitter post failed with status %d: %s", resp.StatusCode, sanitizeErrorBody(responseBody))
	}

	log.Println("[TWITTER] Tweet sent successfully.")
	return nil
}

func buildOAuth1Header(method, rawURL string, bodyParams url.Values, creds TwitterCredentials) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	oauthParams := url.Values{}
	oauthParams.Set("oauth_consumer_key", creds.APIKey)
	oauthParams.Set("oauth_nonce", randomNonce())
	oauthParams.Set("oauth_signature_method", "HMAC-SHA1")
	oauthParams.Set("oauth_timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	oauthParams.Set("oauth_token", creds.AccessToken)
	oauthParams.Set("oauth_version", "1.0")

	signatureBase := oauthSignatureBase(method, parsedURL, mergeValues(parsedURL.Query(), bodyParams, oauthParams))
	signingKey := url.QueryEscape(creds.APISecretKey) + "&" + url.QueryEscape(creds.AccessTokenSecret)
	signature := signHMACSHA1(signingKey, signatureBase)
	oauthParams.Set("oauth_signature", signature)

	var parts []string
	for key, values := range oauthParams {
		for _, value := range values {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, percentEncode(key), percentEncode(value)))
		}
	}
	sort.Strings(parts)
	return "OAuth " + strings.Join(parts, ", "), nil
}

func oauthSignatureBase(method string, parsedURL *url.URL, params url.Values) string {
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	encodedParams := percentEncode(params.Encode())
	return strings.ToUpper(method) + "&" + percentEncode(baseURL) + "&" + encodedParams
}

func mergeValues(values ...url.Values) url.Values {
	merged := url.Values{}
	for _, set := range values {
		for key, vals := range set {
			for _, val := range vals {
				merged.Add(key, val)
			}
		}
	}
	return merged
}

func signHMACSHA1(key, base string) string {
	mac := hmac.New(sha1.New, []byte(key))
	_, _ = mac.Write([]byte(base))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func randomNonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func percentEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

func marshalJSON(payload map[string]string) ([]byte, error) {
	return json.Marshal(payload)
}

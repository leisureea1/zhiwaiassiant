package jwxt

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type JwxtDirectService struct {
	redis *redis.Client
	http  *http.Client
}

func NewJwxtDirectService(redisClient *redis.Client) *JwxtDirectService {
	return &JwxtDirectService{
		redis: redisClient,
		http:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *JwxtDirectService) SessionCacheKey(userID string) string {
	return "jwxt:session:" + userID
}

func (s *JwxtDirectService) SaveSession(ctx context.Context, userID string, sess *CachedJWXTSession, ttl time.Duration) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is nil")
	}
	toSave := *sess
	b, err := json.Marshal(&toSave)
	if err != nil {
		return err
	}

	encrypted, err := encryptSessionPayload(b)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, s.SessionCacheKey(userID), encrypted, ttl).Err()
}

func (s *JwxtDirectService) LoadSession(ctx context.Context, userID string) (*CachedJWXTSession, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	raw, err := s.redis.Get(ctx, s.SessionCacheKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	payload, err := decryptSessionPayload(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session cache, old plaintext format is no longer supported")
	}
	var sess CachedJWXTSession
	if err := json.Unmarshal(payload, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *JwxtDirectService) ClearSession(ctx context.Context, userID string) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Del(ctx, s.SessionCacheKey(userID)).Err()
}

func (s *JwxtDirectService) newSessionClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
		Jar:     jar,
	}
	return client, nil
}

func (s *JwxtDirectService) clientFromSession(sess *CachedJWXTSession) (*http.Client, error) {
	client, err := s.newSessionClient()
	if err != nil {
		return nil, err
	}
	for _, c := range sess.Cookies {
		u, err := url.Parse(c.Raw)
		if err != nil {
			continue
		}
		httpCookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			HttpOnly: c.HttpOnly,
			Secure:   c.Secure,
			MaxAge:   c.MaxAge,
		}
		if c.Expires > 0 {
			httpCookie.Expires = time.Unix(c.Expires, 0)
		}
		client.Jar.SetCookies(u, []*http.Cookie{httpCookie})
	}
	return client, nil
}

func (s *JwxtDirectService) exportSession(client *http.Client, username string) *CachedJWXTSession {
	urls := []string{
		jwxtBaseURL,
		jwxtHomeURL,
		jwxtBaseURL + "/eams/courseTableForStd.action",
		jwxtBaseURL + "/eams/stdDetail.action",
		casLoginURL,
		eamBaseURL,
	}
	cookies := make([]SerializableCookie, 0, 32)
	seen := map[string]struct{}{}
	for _, raw := range urls {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		for _, c := range client.Jar.Cookies(u) {
			key := strings.Join([]string{c.Name, c.Value, c.Domain, c.Path, raw}, "|")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			sc := SerializableCookie{
				Name:     c.Name,
				Value:    c.Value,
				Path:     c.Path,
				Domain:   c.Domain,
				Raw:      raw,
				HttpOnly: c.HttpOnly,
				Secure:   c.Secure,
				MaxAge:   c.MaxAge,
			}
			if !c.Expires.IsZero() {
				sc.Expires = c.Expires.Unix()
			}
			cookies = append(cookies, sc)
		}
	}

	now := time.Now().UnixMilli()
	return &CachedJWXTSession{
		Username:    username,
		Cookies:     cookies,
		ValidatedAt: now,
		CreatedAt:   now,
	}
}

func (s *JwxtDirectService) buildCachedSession(client *http.Client, username, _ string) *CachedJWXTSession {
	sess := s.exportSession(client, username)
	sess.StudentID = strings.TrimSpace(s.getStudentID(client))
	return sess
}

func getSessionSecret() string {
	v := strings.TrimSpace(os.Getenv("JWXT_SESSION_SECRET"))
	if v == "" {
		log.Printf("WARNING: JWXT_SESSION_SECRET not set, JWXT session encryption is disabled")
	}
	return v
}

func encryptSessionPayload(payload []byte) (string, error) {
	secret := getSessionSecret()
	if secret == "" {
		return "", fmt.Errorf("missing session secret")
	}
	key := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nil, nonce, payload, nil)
	blob := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(blob), nil
}

func decryptSessionPayload(raw string) ([]byte, error) {
	secret := getSessionSecret()
	if secret == "" {
		return nil, fmt.Errorf("missing session secret")
	}
	key := sha256.Sum256([]byte(secret))

	blob, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(blob) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid encrypted payload")
	}

	nonce := blob[:gcm.NonceSize()]
	cipherText := blob[gcm.NonceSize():]
	return gcm.Open(nil, nonce, cipherText, nil)
}

func (s *JwxtDirectService) getWithFinalURL(client *http.Client, rawURL string) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", defaultUA)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return string(body), finalURL, nil
}

func (s *JwxtDirectService) get(client *http.Client, rawURL string) (string, error) {
	body, _, _, err := s.getWithMeta(client, rawURL)
	return body, err
}

func (s *JwxtDirectService) getWithMeta(client *http.Client, rawURL string) (string, int, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", 0, nil, err
	}
	req.Header.Set("User-Agent", defaultUA)
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, nil, err
	}
	return string(body), resp.StatusCode, resp.Header.Clone(), nil
}

func (s *JwxtDirectService) getWithMetaNoRedirect(client *http.Client, rawURL string) (string, int, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", 0, nil, err
	}
	req.Header.Set("User-Agent", defaultUA)
	resp, err := s.doNoRedirect(client, req)
	if err != nil {
		return "", 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, nil, err
	}
	return string(body), resp.StatusCode, resp.Header.Clone(), nil
}

func (s *JwxtDirectService) doNoRedirect(client *http.Client, req *http.Request) (*http.Response, error) {
	noRedirectClient := &http.Client{
		Timeout: client.Timeout,
		Jar:     client.Jar,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return noRedirectClient.Do(req)
}

func (s *JwxtDirectService) postForm(client *http.Client, rawURL string, form url.Values) (string, error) {
	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", originOf(rawURL))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *JwxtDirectService) visitIgnoreError(client *http.Client, rawURL string) error {
	_, err := s.get(client, rawURL)
	return err
}

func getCookieValue(client *http.Client, rawURL, name string) string {
	if client == nil || client.Jar == nil {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for _, c := range client.Jar.Cookies(u) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func resolveURL(baseRaw, loc string) string {
	base, err := url.Parse(baseRaw)
	if err != nil {
		return loc
	}
	next, err := url.Parse(loc)
	if err != nil {
		return loc
	}
	return base.ResolveReference(next).String()
}

func originOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func extractTags(html, tag string) []string {
	r := regexp.MustCompile(`(?is)<` + tag + `[^>]*>.*?</` + tag + `>`)
	return r.FindAllString(html, -1)
}

func extractTableHeaders(table string) []string {
	ths := extractTagTexts(table, "th")
	out := make([]string, 0, len(ths))
	for _, th := range ths {
		th = strings.TrimSpace(th)
		if th != "" {
			out = append(out, th)
		}
	}
	return out
}

func extractTableRows(table string) []string {
	r := regexp.MustCompile(`(?is)<tr[^>]*>.*?</tr>`)
	return r.FindAllString(table, -1)
}

func extractTagTexts(html, tag string) []string {
	r := regexp.MustCompile(`(?is)<` + tag + `[^>]*>(.*?)</` + tag + `>`)
	matches := r.FindAllStringSubmatch(html, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, strings.TrimSpace(stripTags(m[1])))
		}
	}
	return out
}

func stripTags(in string) string {
	r := regexp.MustCompile(`(?is)<[^>]+>`)
	s := r.ReplaceAllString(in, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return strings.TrimSpace(s)
}

func parseFloatAny(v any) float64 {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" || s == "<nil>" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func isDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return s != ""
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func parseHTMLTableRows(html string) []map[string]any {
	tables := extractTags(html, "table")
	for _, table := range tables {
		headers := extractTableHeaders(table)
		if len(headers) == 0 {
			continue
		}
		rows := extractTableRows(table)
		out := make([]map[string]any, 0)
		for _, row := range rows {
			cells := extractTagTexts(row, "td")
			// 只处理包含 <td> 的数据行，跳过只有 <th> 的表头行
			if len(cells) == 0 {
				continue
			}
			item := map[string]any{"id": len(out) + 1}
			for i, cell := range cells {
				if i < len(headers) && headers[i] != "" {
					item[headers[i]] = cell
				}
			}
			if len(item) > 1 {
				out = append(out, item)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []map[string]any{}
}

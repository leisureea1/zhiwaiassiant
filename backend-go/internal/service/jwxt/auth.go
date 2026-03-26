package jwxt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func (s *JwxtDirectService) Login(username, password string) (*CachedJWXTSession, error) {
	client, err := s.newSessionClient()
	if err != nil {
		return nil, err
	}

	_ = s.visitIgnoreError(client, portalEntryURL)

	portalService := portalCASRedirect + "?redirect_url=" + url.QueryEscape(portalDefaultTarget)
	portalPageURL := casLoginURL + "?service=" + url.QueryEscape(portalService)
	if page, finalPortalPageURL, err := s.getWithFinalURL(client, portalPageURL); err == nil {
		hidden := extractFormHidden(page)
		if hidden["execution"] != "" {
			_ = s.submitCASLogin(client, finalPortalPageURL, hidden, username, password)
		}
	}

	entryResp, status, headers, err := s.getWithMetaNoRedirect(client, jwxtSSOURL)
	if err != nil {
		return nil, err
	}
	if status == http.StatusOK && strings.Contains(entryResp, "教务管理系统") {
		return s.buildCachedSession(client, username, password), nil
	}

	encodedTarget := "base64" + base64.StdEncoding.EncodeToString([]byte(jwxtHomeURL))
	serviceURL := jwxtSSOURL + "?targetUrl=" + encodedTarget

	casURL := casLoginURL
	if loc := headers.Get("Location"); strings.TrimSpace(loc) != "" {
		casURL = resolveURL(jwxtSSOURL, loc)
	}

	loginPageURL := casURL
	if strings.Contains(casURL, "?") {
		loginPageURL += "&service=" + url.QueryEscape(serviceURL)
	} else {
		loginPageURL += "?service=" + url.QueryEscape(serviceURL)
	}

	loginPage, finalLoginPageURL, err := s.getWithFinalURL(client, loginPageURL)
	if err != nil {
		return nil, err
	}
	if strings.Contains(loginPage, "教务管理系统") {
		return s.buildCachedSession(client, username, password), nil
	}

	hidden := extractFormHidden(loginPage)
	if hidden["execution"] == "" {
		hidden["execution"] = extractExecution(loginPage)
	}
	if hidden["execution"] == "" {
		retryPage, retryFinalURL, err := s.getWithFinalURL(client, loginPageURL)
		if err == nil {
			hidden = extractFormHidden(retryPage)
			finalLoginPageURL = retryFinalURL
			if hidden["execution"] == "" {
				hidden["execution"] = extractExecution(retryPage)
			}
		}
	}
	if hidden["execution"] == "" {
		return nil, fmt.Errorf("failed to get cas execution value")
	}

	if err := s.submitCASLogin(client, finalLoginPageURL, hidden, username, password); err != nil {
		return nil, err
	}

	if _, err := s.get(client, jwxtHomeURL); err != nil {
		return nil, err
	}

	return s.buildCachedSession(client, username, password), nil
}

func (s *JwxtDirectService) ValidateSession(sess *CachedJWXTSession) bool {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return false
	}
	body, err := s.get(client, jwxtHomeURL)
	if err != nil {
		return false
	}
	if strings.Contains(body, "统一身份认证") || strings.Contains(strings.ToLower(body), "cas/login") {
		return false
	}
	return strings.Contains(body, "教务") || strings.Contains(body, "学生") || strings.Contains(body, "课程")
}

func (s *JwxtDirectService) submitCASLogin(client *http.Client, loginPageURL string, hidden map[string]string, username, password string) error {
	encPwd, err := s.encryptCASPassword(client, password)
	if err != nil {
		return err
	}
	v := url.Values{}
	for k, val := range hidden {
		if k != "username" && k != "password" {
			v.Set(k, val)
		}
	}
	if _, ok := hidden["rememberMe"]; ok {
		v.Set("rememberMe", "true")
	}
	v.Set("username", username)
	v.Set("password", encPwd)
	v.Set("_eventId", "submit")
	v.Set("geolocation", "")

	req, err := http.NewRequest(http.MethodPost, loginPageURL, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", defaultUA)

	resp, err := s.doNoRedirect(client, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	text := string(bodyBytes)
	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	if resp.StatusCode == http.StatusUnauthorized || isCASCredentialError(text, finalURL) {
		return fmt.Errorf("用户名或密码错误")
	}
	if isCASCaptchaRequired(text, finalURL, resp.StatusCode) {
		return fmt.Errorf("需要验证码，请稍后重试")
	}

	loc := resp.Header.Get("Location")
	if strings.TrimSpace(loc) != "" {
		ticketURL := resolveURL(loginPageURL, loc)
		_, _, h2, err := s.getWithMetaNoRedirect(client, ticketURL)
		if err == nil {
			if l2 := h2.Get("Location"); strings.TrimSpace(l2) != "" {
				_, _ = s.get(client, resolveURL(ticketURL, l2))
			}
		}
	}

	return nil
}

func isCASCaptchaRequired(html, finalURL string, statusCode int) bool {
	onCASLogin := strings.Contains(finalURL, "/cas/login")
	if !onCASLogin {
		return false
	}
	if statusCode >= 300 && statusCode < 400 {
		return false
	}

	patterns := []string{
		`(?is)needCaptcha\s*[:=]\s*true`,
		`(?is)"needCaptcha"\s*:\s*true`,
		`(?is)请输入验证码`,
		`(?is)验证码错误`,
		`(?is)captcha\s*error`,
	}
	for _, p := range patterns {
		if regexp.MustCompile(p).FindStringIndex(html) != nil {
			return true
		}
	}
	return false
}

func isCASCredentialError(html, finalURL string) bool {
	onCASLogin := strings.Contains(finalURL, "/cas/login")
	if !onCASLogin {
		return false
	}

	patterns := []string{
		`(?is)credentialError`,
		`(?is)用户名或密码错误`,
		`(?is)密码错误`,
		`(?is)invalid\s+credentials`,
		`(?is)<form[^>]*id=["']fm1["']`,
	}
	matched := false
	for _, p := range patterns {
		if regexp.MustCompile(p).FindStringIndex(html) != nil {
			matched = true
			break
		}
	}
	hasFailureHint := regexp.MustCompile(`(?is)(credentialError|用户名或密码错误|密码错误|invalid\s+credentials)`).FindStringIndex(html) != nil
	return matched && hasFailureHint
}

func (s *JwxtDirectService) encryptCASPassword(client *http.Client, password string) (string, error) {
	pemBody, err := s.get(client, casPublicKeyURL)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode([]byte(pemBody))
	if block == nil {
		return "", fmt.Errorf("invalid cas public key pem")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("cas public key is not rsa")
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(password))
	if err != nil {
		return "", err
	}
	return "__RSA__" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func extractExecution(html string) string {
	r := regexp.MustCompile(`(?is)name\s*=\s*["']execution["']\s+[^>]*value\s*=\s*["']([^"']+)["']`)
	m := r.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractFormHidden(html string) map[string]string {
	out := map[string]string{}
	formHTML := html
	fm1Re := regexp.MustCompile(`(?is)<form[^>]*id=["']fm1["'][^>]*>(.*?)</form>`)
	fm1Match := fm1Re.FindStringSubmatch(html)
	if len(fm1Match) > 1 {
		formHTML = fm1Match[1]
	} else {
		formRe := regexp.MustCompile(`(?is)<form[^>]*>(.*?)</form>`)
		formMatch := formRe.FindStringSubmatch(html)
		if len(formMatch) > 1 {
			formHTML = formMatch[1]
		}
	}

	for _, tag := range regexp.MustCompile(`(?is)<input[^>]*>`).FindAllString(formHTML, -1) {
		name := strings.TrimSpace(extractAttr(tag, "name"))
		if name == "" {
			continue
		}
		typeVal := extractAttr(tag, "type")
		if typeVal != "" && strings.ToLower(typeVal) != "hidden" && strings.ToLower(typeVal) != "submit" {
			continue
		}
		out[name] = extractAttr(tag, "value")
	}
	return out
}

func extractAttr(tag, attr string) string {
	quoted := regexp.MustCompile(`(?is)` + regexp.QuoteMeta(attr) + `\s*=\s*["']([^"']*)["']`)
	if m := quoted.FindStringSubmatch(tag); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}

	unquoted := regexp.MustCompile(`(?is)` + regexp.QuoteMeta(attr) + `\s*=\s*([^\s>]+)`)
	if m := unquoted.FindStringSubmatch(tag); len(m) > 1 {
		return strings.TrimSpace(strings.Trim(m[1], `"'`))
	}
	return ""
}

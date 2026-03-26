package jwxt

const (
	casLoginURL     = "https://login.xisu.edu.cn/cas/login"
	casPublicKeyURL = "https://login.xisu.edu.cn/cas/jwt/publicKey"

	jwxtBaseURL = "https://jwxt.xisu.edu.cn"
	jwxtHomeURL = "https://jwxt.xisu.edu.cn/eams/home.action"
	jwxtSSOURL  = "https://jwxt.xisu.edu.cn/eams/sso/login.action"

	portalEntryURL      = "https://wsbsdt.xisu.edu.cn/page/site/index"
	portalCASRedirect   = "https://wsbsdt.xisu.edu.cn/common/actionCasLogin"
	portalDefaultTarget = "https://wsbsdt.xisu.edu.cn/page/site/visitor"

	eamBaseURL  = "http://59.74.65.32:8080"
	eamSSOEntry = "http://59.74.65.32:8080/eamssoshuwei/sso/login.jsp"

	defaultUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36"
)

type SerializableCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Raw      string `json:"raw"`
	Expires  int64  `json:"expires"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	MaxAge   int    `json:"maxAge"`
}

type CachedJWXTSession struct {
	Username    string               `json:"username"`
	Cookies     []SerializableCookie `json:"cookies"`
	StudentID   string               `json:"studentId,omitempty"`
	ValidatedAt int64                `json:"validatedAt"`
	CreatedAt   int64                `json:"createdAt"`
	EAMAuthed   bool                 `json:"eamAuthed"`
}

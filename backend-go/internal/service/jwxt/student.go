package jwxt

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func (s *JwxtDirectService) GetUser(sess *CachedJWXTSession) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	studentID := s.getStudentID(client)
	if strings.TrimSpace(studentID) == "" {
		studentID = strings.TrimSpace(sess.StudentID)
	}
	detail := map[string]any{
		"name":         nil,
		"student_code": nil,
		"department":   nil,
		"major":        nil,
		"class_name":   nil,
		"grade":        nil,
	}

	if detailPage, e := s.get(client, jwxtBaseURL+"/eams/stdDetail.action"); e == nil {
		mergeUserDetail(detail, detailPage)
	}

	weekInfo := map[string]any{
		"current_week":  nil,
		"semester_name": nil,
	}
	if weekPage, e := s.get(client, jwxtBaseURL+"/eams/home!welcome.action"); e == nil {
		mergeWeekInfo(weekInfo, weekPage)
	}

	info := map[string]any{
		"success":    true,
		"student_id": strings.TrimSpace(studentID),
	}
	for k, v := range detail {
		info[k] = v
	}
	for k, v := range weekInfo {
		info[k] = v
	}

	if current := getCookieValue(client, jwxtHomeURL, "semester.id"); strings.TrimSpace(current) != "" {
		info["current_semester"] = strings.TrimSpace(current)
	} else if current := getCookieValue(client, jwxtBaseURL+"/eams/courseTableForStd.action", "semester.id"); strings.TrimSpace(current) != "" {
		info["current_semester"] = strings.TrimSpace(current)
	}

	return info, nil
}

func (s *JwxtDirectService) getStudentID(client *http.Client) string {
	patterns := []string{
		`bg\.form\.addInput\s*\(\s*form\s*,\s*["']ids["']\s*,\s*["'](\d+)["']\s*\)`,
		`["']?ids["']?\s*[:=]\s*["'](\d+)["']`,
		`["']?ids["']?\s*[:=]\s*(\d+)`,
		`name\s*=\s*"ids"\s+value\s*=\s*"(\d+)"`,
		`addInput\([^)]*"ids"[^)]*"(\d+)"\)`,
		`[?&]ids=(\d+)`,
		`[?&]student\.id=(\d+)`,
		`["']?student\.id["']?\s*[:=]\s*["']?(\d+)`,
		`["']?std\.id["']?\s*[:=]\s*["']?(\d+)`,
	}
	urls := []string{
		jwxtBaseURL + "/eams/courseTableForStd.action",
		jwxtBaseURL + "/eams/courseTableForStd!index.action",
		jwxtBaseURL + "/eams/stdDetail.action",
		jwxtBaseURL + "/eams/home.action",
	}
	for _, u := range urls {
		text, finalURL, err := s.getWithFinalURL(client, u)
		if err != nil {
			continue
		}
		target := text + "\n" + finalURL
		for _, p := range patterns {
			r := regexp.MustCompile("(?i)" + p)
			m := r.FindStringSubmatch(target)
			if len(m) > 1 {
				return m[1]
			}
		}
	}
	return ""
}

func (s *JwxtDirectService) getCurrentSemesterID(client *http.Client) string {
	if ck := getCookieValue(client, jwxtHomeURL, "semester.id"); strings.TrimSpace(ck) != "" {
		return strings.TrimSpace(ck)
	}
	if ck := getCookieValue(client, jwxtBaseURL+"/eams/courseTableForStd.action", "semester.id"); strings.TrimSpace(ck) != "" {
		return strings.TrimSpace(ck)
	}

	// 优先通过 dataQuery 学期下拉解析，避免页面脚本中的无关数字误匹配。
	form := url.Values{}
	form.Set("dataType", "semester")
	if text, err := s.postForm(client, jwxtBaseURL+"/eams/dataQuery.action", form); err == nil {
		if id := pickSemesterID(extractSemesterOptions(text)); id != "" {
			return id
		}
	}

	urls := []string{
		jwxtBaseURL + "/eams/courseTableForStd.action",
		jwxtBaseURL + "/eams/home.action",
		jwxtBaseURL + "/eams/teach/grade/course/person!search.action",
	}
	patterns := []string{
		`semester\.id["']?\s*[:=]\s*["']?(\d{2,4})`,
		`semesterId["']?\s*[:=]\s*["']?(\d{2,4})`,
		`id="semester"[^>]*value="(\d{2,4})"`,
		`name="semester\.id"[^>]*value="(\d{2,4})"`,
	}
	for _, u := range urls {
		text, err := s.get(client, u)
		if err != nil {
			continue
		}
		if id := pickSemesterID(extractSemesterOptions(text)); id != "" {
			return id
		}
		for _, p := range patterns {
			r := regexp.MustCompile(`(?is)` + p)
			m := r.FindStringSubmatch(text)
			if len(m) > 1 {
				return m[1]
			}
		}
	}
	return ""
}

func mergeUserDetail(info map[string]any, html string) {
	labels := map[string]string{
		"姓名":      "name",
		"学号":      "student_code",
		"院系":      "department",
		"专业":      "major",
		"所属班级":    "class_name",
		"年级":      "grade",
	}
	tableRe := regexp.MustCompile(`(?is)<table[^>]*>.*?</table>`)
	rows := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`).FindAllStringSubmatch(html, -1)
	cellRe := regexp.MustCompile(`(?is)<(td|th)[^>]*>(.*?)</(td|th)>`)
	if len(tableRe.FindAllString(html, -1)) == 0 {
		rows = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`).FindAllStringSubmatch(html, -1)
	}
	for _, row := range rows {
		cells := make([]string, 0, 8)
		for _, cm := range cellRe.FindAllStringSubmatch(row[1], -1) {
			if len(cm) < 3 {
				continue
			}
			cells = append(cells, strings.TrimSpace(stripTags(cm[2])))
		}
		for i := 0; i+1 < len(cells); i++ {
			label := normalizeDetailLabel(cells[i])
			if key, ok := labels[label]; ok && strings.TrimSpace(fmt.Sprintf("%v", info[key])) == "<nil>" {
				value := strings.TrimSpace(cells[i+1])
				if value != "" {
					info[key] = value
				}
			}
		}
	}
}

func normalizeDetailLabel(label string) string {
	label = strings.TrimSpace(stripTags(label))
	label = strings.TrimRight(label, ":：")
	label = strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(label)
	return label
}

func mergeWeekInfo(info map[string]any, html string) {
	if m := regexp.MustCompile(`第\s*(\d+)\s*周`).FindStringSubmatch(html); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			info["current_week"] = n
		}
	}
	if m := regexp.MustCompile(`(\d{4}[-~–]\d{4})学年.*?第\s*(\d+)\s*学期`).FindStringSubmatch(html); len(m) > 2 {
		info["semester_name"] = fmt.Sprintf("%s学年第%s学期", m[1], m[2])
	}
}

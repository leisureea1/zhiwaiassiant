package jwxt

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func (s *JwxtDirectService) GetSemester(sess *CachedJWXTSession) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("dataType", "semester")
	body, err := s.postForm(client, jwxtBaseURL+"/eams/dataQuery.action", form)
	if err != nil {
		return nil, err
	}

	options := extractSemesterOptions(body)
	if len(options) == 0 {
		fallbackURLs := []string{
			jwxtBaseURL + "/eams/courseTableForStd.action",
			jwxtBaseURL + "/eams/home.action",
			jwxtBaseURL + "/eams/teach/grade/course/person!search.action",
		}
		for _, u := range fallbackURLs {
			if page, e := s.get(client, u); e == nil {
				options = extractSemesterOptions(page)
				if len(options) > 0 {
					break
				}
			}
		}
	}
	if len(options) == 0 {
		current := s.getCurrentSemesterID(client)
		if strings.TrimSpace(current) != "" {
			return map[string]any{
				"success":             true,
				"current_semester_id": current,
				"semesters":           []map[string]any{{"id": current, "name": current, "current": true}},
			}, nil
		}
		return map[string]any{"success": false, "error": "获取学期失败", "semesters": []any{}}, nil
	}

	// 与旧 Nest/Python 行为保持一致：按抓取顺序反转，最新学期在前
	for i, j := 0, len(options)-1; i < j; i, j = i+1, j-1 {
		options[i], options[j] = options[j], options[i]
	}

	current := ""
	currentName := s.getCurrentSemesterName(client)

	for _, sem := range options {
		name := strings.TrimSpace(sem["name"].(string))
		if normalizeSemesterText(name) == normalizeSemesterText(currentName) && currentName != "" {
			sem["current"] = true
			current = sem["id"].(string)
		} else {
			sem["current"] = false
		}
	}

	if current == "" {
		current = s.getCurrentSemesterID(client)
		for _, sem := range options {
			if sem["id"] == current {
				sem["current"] = true
			}
		}
	}

	return map[string]any{
		"success":             true,
		"current_semester_id": current,
		"semesters":           options,
	}, nil
}

func extractSemesterOptions(html string) []map[string]any {
	optionTagRe := regexp.MustCompile(`(?is)<option\b[^>]*>.*?</option>`)
	tags := optionTagRe.FindAllString(html, -1)
	out := make([]map[string]any, 0, len(tags))
	for _, tag := range tags {
		id := extractAttr(tag, "value")
		if id == "" {
			if m := regexp.MustCompile(`(?is)\bvalue\s*=\s*([0-9]+)`).FindStringSubmatch(tag); len(m) > 1 {
				id = strings.TrimSpace(m[1])
			}
		}
		if !regexp.MustCompile(`^\d+$`).MatchString(strings.TrimSpace(id)) {
			continue
		}

		textMatch := regexp.MustCompile(`(?is)<option\b[^>]*>(.*?)</option>`).FindStringSubmatch(tag)
		if len(textMatch) < 2 {
			continue
		}
		name := strings.TrimSpace(stripTags(textMatch[1]))
		item := map[string]any{"id": strings.TrimSpace(id), "name": name}
		out = append(out, item)
	}
	return out
}

func (s *JwxtDirectService) getCurrentSemesterName(client *http.Client) string {
	body, err := s.get(client, jwxtBaseURL+"/eams/home.action")
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`(\d{4}[-—]\d{4})\s*学年\s*第?\s*([12])\s*学期`)
	m := re.FindStringSubmatch(body)

	if len(m) < 3 {
		return ""
	}

	year := strings.ReplaceAll(m[1], "—", "-")
	return fmt.Sprintf("%s-%s", year, m[2])
}

func normalizeSemesterText(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`(\d{4}[-—]\d{4})`)
	year := re.FindString(s)
	year = strings.ReplaceAll(year, "—", "-")

	term := ""
	if strings.Contains(s, "第1学期") || strings.HasSuffix(s, "-1") || strings.Contains(s, "1学期") {
		term = "1"
	} else if strings.Contains(s, "第2学期") || strings.HasSuffix(s, "-2") || strings.Contains(s, "2学期") {
		term = "2"
	}

	if year == "" || term == "" {
		return s
	}

	return fmt.Sprintf("%s-%s", year, term)
}

package jwxt

import (
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
	for _, sem := range options {
		if v, ok := sem["current"].(bool); ok && v {
			current = sem["id"].(string)
			break
		}
	}
	if current == "" {
		current = s.getCurrentSemesterID(client)
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
		attrsLower := strings.ToLower(tag)
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
		if strings.Contains(attrsLower, "selected") {
			item["current"] = true
		}
		out = append(out, item)
	}
	return out
}

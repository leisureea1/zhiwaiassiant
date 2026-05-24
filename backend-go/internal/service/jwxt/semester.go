package jwxt

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	fallbackURLs := []string{
		jwxtBaseURL + "/eams/courseTableForStd.action",
		jwxtBaseURL + "/eams/home.action",
		jwxtBaseURL + "/eams/teach/grade/course/person!search.action",
	}
	for _, u := range fallbackURLs {
		if page, e := s.get(client, u); e == nil {
			options = mergeSemesterOptions(options, extractSemesterOptions(page))
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
	currentName, currentWeek := s.getCurrentSemesterInfo(client)

	for _, sem := range options {
		name := strings.TrimSpace(sem["name"].(string))
		if normalizeSemesterText(name) == normalizeSemesterText(currentName) && currentName != "" {
			sem["current"] = true
			current = sem["id"].(string)
		} else {
			sem["current"] = false
		}
	}
	if current == "" && currentName != "" {
		if inferredID := inferSemesterIDFromOptions(currentName, options); inferredID != "" {
			current = inferredID
			options = append([]map[string]any{{
				"id":      inferredID,
				"name":    formatSemesterDisplayName(currentName),
				"current": true,
			}}, options...)
		}
	}

	return map[string]any{
		"success":             true,
		"current_semester_id": current,
		"current_week":        currentWeek,
		"semesters":           options,
	}, nil
}

func mergeSemesterOptions(base, extra []map[string]any) []map[string]any {
	seen := make(map[string]bool, len(base)+len(extra))
	out := make([]map[string]any, 0, len(base)+len(extra))

	for _, group := range [][]map[string]any{base, extra} {
		for _, sem := range group {
			id := strings.TrimSpace(fmt.Sprintf("%v", sem["id"]))
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, sem)
		}
	}

	return out
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
		if normalizeSemesterText(name) == "" {
			continue
		}
		item := map[string]any{"id": strings.TrimSpace(id), "name": name}
		out = append(out, item)
	}
	return out
}

func (s *JwxtDirectService) getCurrentSemesterInfo(client *http.Client) (string, int) {
	url := fmt.Sprintf("%s/eams/home!welcome.action?_=%d", jwxtBaseURL, time.Now().UnixMilli())
	body, err := s.get(client, url)
	if err != nil {
		return "", 0
	}

	name := extractCurrentSemesterNameFromHTML(body)
	week := extractCurrentWeekFromHTML(body)
	return name, week
}

func extractCurrentWeekFromHTML(body string) int {
	// 匹配 "第13周"、"第13教学周"、"第 13 教学周" 等各种变体
	re := regexp.MustCompile(`第\s*(\d+)\s*(?:教学)?周`)
	if m := re.FindStringSubmatch(body); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	text := html.UnescapeString(stripTags(body))
	if m := re.FindStringSubmatch(text); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	return 0
}

func extractCurrentSemesterNameFromHTML(body string) string {
	text := html.UnescapeString(stripTags(body))

	re := regexp.MustCompile(`(\d{4}-\d{4}学年[第]?[12一二][学期]*)`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}

	return normalizeSemesterText(m[1])
}

func normalizeSemesterTerm(term string) string {
	switch strings.TrimSpace(term) {
	case "1", "一":
		return "1"
	case "2", "二":
		return "2"
	default:
		return ""
	}
}

func normalizeSemesterText(s string) string {
	s = html.UnescapeString(strings.TrimSpace(s))
	s = strings.NewReplacer(
		"\u00a0", "",
		" ", "",
		"\t", "",
		"\r", "",
		"\n", "",
		"学年", "-",
		"第", "",
		"学期", "",
		"\u2014", "-",
		"\u2013", "-",
		"\uff0d", "-",
		"\uff5e", "-",
		"~", "-",
	).Replace(s)

	s = strings.ReplaceAll(s, "一", "1")
	s = strings.ReplaceAll(s, "二", "2")

	re := regexp.MustCompile(`(\d{4})-(\d{4})-?([12])`)
	m := re.FindStringSubmatch(s)
	if len(m) < 4 {
		return ""
	}

	return fmt.Sprintf("%s-%s-%s", m[1], m[2], m[3])
}

func inferSemesterIDFromOptions(currentName string, options []map[string]any) string {
	currentStartYear, currentTerm, ok := parseNormalizedSemester(currentName)
	if !ok {
		return ""
	}
	currentSeq := semesterSequence(currentStartYear, currentTerm)

	for _, sem := range options {
		idText := strings.TrimSpace(fmt.Sprintf("%v", sem["id"]))
		id, ok := parseInt(idText)
		if !ok {
			continue
		}
		name := strings.TrimSpace(fmt.Sprintf("%v", sem["name"]))
		startYear, term, ok := parseNormalizedSemester(name)
		if !ok {
			continue
		}
		inferred := id + currentSeq - semesterSequence(startYear, term)
		if inferred > 0 {
			return fmt.Sprintf("%d", inferred)
		}
	}

	return ""
}

func parseNormalizedSemester(s string) (int, int, bool) {
	normalized := normalizeSemesterText(s)
	if normalized == "" {
		return 0, 0, false
	}

	m := regexp.MustCompile(`^(\d{4})-\d{4}-([12])$`).FindStringSubmatch(normalized)
	if len(m) < 3 {
		return 0, 0, false
	}

	startYear, ok := parseInt(m[1])
	if !ok {
		return 0, 0, false
	}
	term, ok := parseInt(m[2])
	if !ok || term < 1 || term > 2 {
		return 0, 0, false
	}

	return startYear, term, true
}

func semesterSequence(startYear, term int) int {
	return startYear*2 + term - 1
}

func formatSemesterDisplayName(normalized string) string {
	startYear, term, ok := parseNormalizedSemester(normalized)
	if !ok {
		return normalized
	}
	termText := "第一"
	if term == 2 {
		termText = "第二"
	}
	return fmt.Sprintf("%d-%d学年第%s学期", startYear, startYear+1, termText)
}

func parseInt(s string) (int, bool) {
	n := 0
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return 0, false
	}
	return n, true
}

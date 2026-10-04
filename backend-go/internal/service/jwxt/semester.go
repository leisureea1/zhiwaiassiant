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
		return map[string]any{"success": false, "error": "获取学期失败", "semesters": []map[string]any{}}, fmt.Errorf("failed to fetch semesters from jwxt: no semester options found")
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

	// Fallback 1: check if any option had `selected` in HTML
	if current == "" {
		for _, sem := range options {
			if sel, ok := sem["selected"].(bool); ok && sel {
				sem["current"] = true
				current = fmt.Sprintf("%v", sem["id"])
				break
			}
		}
	}

	// Fallback 2: try getCurrentSemesterID (cookies, scraping action pages)
	if current == "" {
		if curID := s.getCurrentSemesterID(client); strings.TrimSpace(curID) != "" {
			current = strings.TrimSpace(curID)
			found := false
			for _, sem := range options {
				if fmt.Sprintf("%v", sem["id"]) == current {
					sem["current"] = true
					found = true
					break
				}
			}
			if !found {
				options = append([]map[string]any{{
					"id":      current,
					"name":    current,
					"current": true,
				}}, options...)
			}
		}
	}

	// Fallback 3: fallback to latest semester in options
	if current == "" && len(options) > 0 {
		options[0]["current"] = true
		current = fmt.Sprintf("%v", options[0]["id"])
	}

	// Fallback current_week if not extracted from HTML
	if currentWeek <= 0 && currentName != "" {
		currentWeek = inferCurrentWeekFromSemester(currentName, time.Now())
	}
	if currentWeek <= 0 {
		for _, sem := range options {
			if isCur, ok := sem["current"].(bool); ok && isCur {
				if name, ok := sem["name"].(string); ok {
					currentWeek = inferCurrentWeekFromSemester(name, time.Now())
					if currentWeek > 0 {
						break
					}
				}
			}
		}
	}
	if currentWeek <= 0 && len(options) > 0 {
		if name, ok := options[0]["name"].(string); ok {
			currentWeek = inferCurrentWeekFromSemester(name, time.Now())
		}
	}
	if currentWeek <= 0 {
		currentWeek = 1
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
		if regexp.MustCompile(`(?i)\bselected\b`).MatchString(tag) {
			item["selected"] = true
		}
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
	patterns := []string{
		`第\s*(\d+)\s*(?:教学)?周`,
		`(?:教学周|当前教学周|当前周|周次)[：:\s]+(?:第)?\s*(\d+)`,
		`["']?(?:curWeek|currentWeek|teachWeek|weekNumber)["']?\s*[:=]\s*["']?(\d+)`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil && n > 0 && n <= 30 {
				return n
			}
		}
	}

	text := html.UnescapeString(stripTags(body))
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(text); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil && n > 0 && n <= 30 {
				return n
			}
		}
	}
	return 0
}

func inferCurrentWeekFromSemester(semesterName string, now time.Time) int {
	startYear, term, ok := parseNormalizedSemester(semesterName)
	if !ok {
		return 0
	}
	loc := time.FixedZone("CST", 8*3600) // 中国标准时间 UTC+8
	nowCST := now.In(loc)
	var semStart time.Time

	if term == 1 {
		// 秋季学期：9月1日所在周的周一
		sep1 := time.Date(startYear, time.September, 1, 0, 0, 0, 0, loc)
		weekday := int(sep1.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		semStart = sep1.AddDate(0, 0, -(weekday - 1))
	} else if term == 2 {
		// 春季学期：次年3月1日所在周的周一
		mar1 := time.Date(startYear+1, time.March, 1, 0, 0, 0, 0, loc)
		weekday := int(mar1.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		semStart = mar1.AddDate(0, 0, -(weekday - 1))
	} else {
		return 0
	}

	if nowCST.Before(semStart) {
		return 1
	}

	diffDays := int(nowCST.Sub(semStart).Hours() / 24)
	week := diffDays/7 + 1
	if week >= 1 && week <= 25 {
		return week
	}
	return 0
}

func extractCurrentSemesterNameFromHTML(body string) string {
	text := html.UnescapeString(stripTags(body))

	patterns := []string{
		`(\d{4}\s*[-–—~～]\s*\d{4}\s*(?:学年|学年度|年度)?\s*(?:第)?\s*[123一二三两]\s*学期)`,
		`(\d{4}\s*[-–—~～]\s*\d{4}\s*(?:学年|学年度|年度)\s*(?:第)?\s*[123一二三两])`,
		`(\d{4}\s*[-–—~～]\s*\d{4}\s*[-_]\s*[123])`,
		`(\d{4}\s*[-–—~～]\s*\d{4}\s*(?:学年|学年度|年度)?\s*(?:春季|秋季|夏季|秋|春|夏)\s*(?:学期)?)`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(`(?i)` + p)
		if m := re.FindStringSubmatch(text); len(m) > 1 {
			if norm := normalizeSemesterText(m[1]); norm != "" {
				return norm
			}
		}
	}

	return ""
}

func normalizeSemesterTerm(term string) string {
	switch strings.TrimSpace(term) {
	case "1", "一":
		return "1"
	case "2", "二", "两":
		return "2"
	case "3", "三":
		return "3"
	default:
		return ""
	}
}

func normalizeTermDigit(t string) string {
	switch strings.TrimSpace(t) {
	case "1", "一":
		return "1"
	case "2", "二", "两":
		return "2"
	case "3", "三":
		return "3"
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
		"\u2014", "-",
		"\u2013", "-",
		"\uff0d", "-",
		"\uff5e", "-",
		"~", "-",
	).Replace(s)

	yearRe := regexp.MustCompile(`(\d{4})-(\d{4})`)
	mYear := yearRe.FindStringSubmatchIndex(s)
	if len(mYear) < 4 {
		return ""
	}
	y1 := s[mYear[2]:mYear[3]]
	y2 := s[mYear[4]:mYear[5]]

	rest := s[mYear[1]:]

	term := ""
	// 1. Direct dash-term, e.g. -1, -2, -3
	if m := regexp.MustCompile(`^-([123])(?:\D|$)`).FindStringSubmatch(rest); len(m) > 1 {
		term = m[1]
	}

	// 2. Chinese term patterns
	if term == "" {
		if m := regexp.MustCompile(`(?:第)?([123一二三两])\s*学期`).FindStringSubmatch(rest); len(m) > 1 {
			term = normalizeTermDigit(m[1])
		}
	}

	if term == "" {
		if m := regexp.MustCompile(`(?:学年|学年度|年度)\s*(?:第)?\s*([123一二三两])`).FindStringSubmatch(rest); len(m) > 1 {
			term = normalizeTermDigit(m[1])
		}
	}

	if term == "" {
		if strings.Contains(rest, "秋") {
			term = "1"
		} else if strings.Contains(rest, "春") {
			term = "2"
		} else if strings.Contains(rest, "夏") || strings.Contains(rest, "短") {
			term = "3"
		}
	}

	if term == "" {
		return ""
	}

	return fmt.Sprintf("%s-%s-%s", y1, y2, term)
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

	m := regexp.MustCompile(`^(\d{4})-\d{4}-([123])$`).FindStringSubmatch(normalized)
	if len(m) < 3 {
		return 0, 0, false
	}

	startYear, ok := parseInt(m[1])
	if !ok {
		return 0, 0, false
	}
	term, ok := parseInt(m[2])
	if !ok || term < 1 || term > 3 {
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
	} else if term == 3 {
		termText = "第三"
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

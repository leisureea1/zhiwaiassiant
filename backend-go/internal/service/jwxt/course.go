package jwxt

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func (s *JwxtDirectService) GetCourse(sess *CachedJWXTSession, semesterID, studentID string) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	_, _ = s.get(client, jwxtBaseURL+"/eams/courseTableForStd.action")

	if strings.TrimSpace(semesterID) == "" {
		semesterID = s.getCurrentSemesterID(client)
	}
	if strings.TrimSpace(semesterID) == "" {
		semesterID = s.getLatestSemesterID(client)
	}
	if strings.TrimSpace(semesterID) == "" {
		return map[string]any{"success": false, "error": "无法获取当前学期ID", "courses": []any{}}, nil
	}
	if strings.TrimSpace(studentID) == "" {
		studentID = s.getStudentID(client)
	}
	if strings.TrimSpace(studentID) == "" {
		studentID = strings.TrimSpace(sess.StudentID)
	}
	if strings.TrimSpace(studentID) == "" {
		return map[string]any{"success": false, "error": "无法获取学生ID", "courses": []any{}}, nil
	}

	form := url.Values{}
	form.Set("ignoreHead", "1")
	form.Set("setting.kind", "std")
	form.Set("startWeek", "")
	form.Set("semester.id", semesterID)
	form.Set("ids", studentID)

	body, err := s.postForm(client, jwxtBaseURL+"/eams/courseTableForStd!courseTable.action", form)
	if err != nil {
		return nil, err
	}

	courses := parseCourseActivities(body)
	return map[string]any{
		"success":       true,
		"courses":       courses,
		"total_courses": len(courses),
		"semester_id":   semesterID,
	}, nil
}

func (s *JwxtDirectService) getLatestSemesterID(client *http.Client) string {
	form := url.Values{}
	form.Set("dataType", "semester")
	if body, err := s.postForm(client, jwxtBaseURL+"/eams/dataQuery.action", form); err == nil {
		if id := pickSemesterID(extractSemesterOptions(body)); id != "" {
			return id
		}
	}

	fallbackURLs := []string{
		jwxtBaseURL + "/eams/courseTableForStd.action",
		jwxtBaseURL + "/eams/home.action",
		jwxtBaseURL + "/eams/teach/grade/course/person!search.action",
	}
	for _, raw := range fallbackURLs {
		if page, err := s.get(client, raw); err == nil {
			if id := pickSemesterID(extractSemesterOptions(page)); id != "" {
				return id
			}
		}
	}

	return ""
}

func pickSemesterID(options []map[string]any) string {
	if len(options) == 0 {
		return ""
	}

	for _, sem := range options {
		if current, ok := sem["current"].(bool); ok && current {
			if id, ok := sem["id"].(string); ok {
				return strings.TrimSpace(id)
			}
		}
	}

	bestID := ""
	bestNum := -1
	for _, sem := range options {
		id, ok := sem["id"].(string)
		if !ok {
			continue
		}
		id = strings.TrimSpace(id)
		n, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		if n > bestNum {
			bestNum = n
			bestID = id
		}
	}

	return bestID
}

func parseCourseActivities(html string) []map[string]any {
	results := make([]map[string]any, 0)
	parts := strings.Split(html, "var teachers")
	activityRe := regexp.MustCompile(`(?is)activity\s*=\s*new\s+TaskActivity\s*\((.*?)\);(.*)`)
	for i := 1; i < len(parts); i++ {
		block := "var teachers" + parts[i]
		m := activityRe.FindStringSubmatch(block)
		if len(m) < 3 {
			continue
		}

		params := splitCSVArgs(m[1])
		if len(params) < 7 {
			continue
		}
		teacher := extractTeacherName(block)
		if teacher == "" {
			teacher = cleanQuoted(params[1])
		}
		rawCourseName := cleanQuoted(params[3])
		courseName := strings.TrimSpace(strings.Split(rawCourseName, "(")[0])
		if shouldSkipCourseName(courseName) {
			continue
		}
		classroom := cleanQuoted(params[5])
		if isOnlineCourse(rawCourseName, courseName, classroom) {
			continue
		}
		timePattern := cleanQuoted(params[6])

		timeSlots := parseTimeSlots(m[2])
		weeks := parseWeeks(timePattern)

		dayPeriods := map[int][]int{}
		for _, slot := range timeSlots {
			day := slot["weekday_index"].(int)
			period := slot["period"].(int)
			dayPeriods[day] = append(dayPeriods[day], period)
		}

		for day, periods := range dayPeriods {
			sort.Ints(periods)
			if len(periods) == 0 {
				continue
			}
			results = append(results, map[string]any{
				"name":         courseName,
				"teacher":      teacher,
				"classroom":    classroom,
				"weekday":      day,
				"startSection": periods[0],
				"endSection":   periods[len(periods)-1],
				"weeks":        formatWeeks(weeks),
			})
		}
	}
	return results
}

func shouldSkipCourseName(name string) bool {
	n := strings.TrimSpace(strings.ToLower(name))
	if n == "" {
		return true
	}

	replacer := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", "：", "", ":", "")
	normalized := replacer.Replace(n)

	// 过滤教务页面表头/占位文案，避免被误识别为课程
	invalid := map[string]struct{}{
		"课程名称":     {},
		"课程名称课程代码": {},
		"课程名":      {},
		"课程":        {},
		"course":    {},
		"coursename": {},
	}
	if _, ok := invalid[normalized]; ok {
		return true
	}

	if strings.Contains(normalized, "课程名称") || strings.Contains(normalized, "coursename") {
		return true
	}

	return false
}

func isOnlineCourse(rawName, name, classroom string) bool {
	fields := []string{rawName, name, classroom}
	for _, field := range fields {
		n := normalizeCourseText(field)
		if n == "" {
			continue
		}

		markers := []string{
			"网上课程",
			"网络课程",
			"线上课程",
			"在线课程",
			"网课",
			"慕课",
			"mooc",
			"online",
		}
		for _, marker := range markers {
			if strings.Contains(n, marker) {
				return true
			}
		}
	}

	return false
}

func normalizeCourseText(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\r", "",
		"\n", "",
		"（", "(",
		"）", ")",
		"：", ":",
	)
	return replacer.Replace(s)
}

func splitCSVArgs(in string) []string {
	out := make([]string, 0, 16)
	var cur strings.Builder
	inQuotes := false
	var q rune
	for _, c := range in {
		if c == '\'' || c == '"' {
			if !inQuotes {
				inQuotes = true
				q = c
			} else if c == q {
				inQuotes = false
			}
			cur.WriteRune(c)
			continue
		}
		if c == ',' && !inQuotes {
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteRune(c)
	}
	if strings.TrimSpace(cur.String()) != "" {
		out = append(out, strings.TrimSpace(cur.String()))
	}
	return out
}

func cleanQuoted(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func extractTeacherName(code string) string {
	r := regexp.MustCompile(`(?is)var\s+actTeachers\s*=\s*\[(.*?)\];`)
	m := r.FindStringSubmatch(code)
	if len(m) < 2 {
		return ""
	}
	r2 := regexp.MustCompile(`(?is)name\s*:\s*["']([^"']+)["']`)
	ms := r2.FindAllStringSubmatch(m[1], -1)
	names := make([]string, 0, len(ms))
	for _, item := range ms {
		if len(item) > 1 {
			names = append(names, strings.TrimSpace(item[1]))
		}
	}
	return strings.Join(names, ", ")
}

func parseTimeSlots(code string) []map[string]any {
	weekdayNames := []string{"星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"}

	slots := make([]map[string]any, 0)
	rA := regexp.MustCompile(`index\s*=\s*(\d+)\s*\*\s*unitCount\s*\+\s*(\d+)\s*;`)
	rB := regexp.MustCompile(`index\s*=\s*(\d+)\s*;`)

	for _, m := range rA.FindAllStringSubmatch(code, -1) {
		if len(m) < 3 {
			continue
		}
		a, _ := strconv.Atoi(m[1])
		b, _ := strconv.Atoi(m[2])
		idx := a*12 + b
		weekdayIdx := idx / 12
		periodIdx := idx % 12
		if weekdayIdx < 0 || weekdayIdx >= 7 {
			continue
		}
		slots = append(slots, map[string]any{
			"weekday":       weekdayNames[weekdayIdx],
			"weekday_index": weekdayIdx + 1,
			"period":        periodIdx + 1,
		})
	}

	if len(slots) == 0 {
		for _, m := range rB.FindAllStringSubmatch(code, -1) {
			if len(m) < 2 {
				continue
			}
			idx, _ := strconv.Atoi(m[1])
			weekdayIdx := idx / 12
			periodIdx := idx % 12
			if weekdayIdx < 0 || weekdayIdx >= 7 {
				continue
			}
			slots = append(slots, map[string]any{
				"weekday":       weekdayNames[weekdayIdx],
				"weekday_index": weekdayIdx + 1,
				"period":        periodIdx + 1,
			})
		}
	}

	return slots
}

func parseWeeks(pattern string) []int {
	weeks := make([]int, 0)
	if len(pattern) <= 10 {
		return weeks
	}
	for i, c := range pattern {
		if c == '1' && i > 0 {
			weeks = append(weeks, i)
		}
	}
	return weeks
}

func formatWeeks(weeks []int) string {
	if len(weeks) == 0 {
		return ""
	}
	parts := make([]string, 0)
	start := weeks[0]
	end := weeks[0]
	for _, w := range weeks[1:] {
		if w == end+1 {
			end = w
			continue
		}
		if start == end {
			parts = append(parts, fmt.Sprintf("%d周", start))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d周", start, end))
		}
		start = w
		end = w
	}
	if start == end {
		parts = append(parts, fmt.Sprintf("%d周", start))
	} else {
		parts = append(parts, fmt.Sprintf("%d-%d周", start, end))
	}
	return strings.Join(parts, ", ")
}

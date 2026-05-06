package jwxt

import (
	"fmt"
	"net/url"
	"strings"
)

func (s *JwxtDirectService) GetExam(sess *CachedJWXTSession, semesterID string) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	_, _ = s.get(client, jwxtBaseURL+"/eams/stdExamTable.action")

	semesterID = strings.TrimSpace(semesterID)
	examURL := jwxtBaseURL + "/eams/stdExamTable!examTable.action"
	if semesterID != "" {
		q := url.Values{}
		q.Set("semester.id", semesterID)
		examURL += "?" + q.Encode()
	}
	body, err := s.get(client, examURL)
	if err != nil {
		return nil, err
	}
	if isJWXTLoginPage(body) {
		return map[string]any{
			"success":  false,
			"error":    "需要重新登录教务系统",
			"exams":    []map[string]any{},
			"semester": semesterID,
		}, nil
	}

	exams := parseExamRows(body)
	return map[string]any{"success": true, "exams": exams, "semester": semesterID}, nil
}

func parseExamRows(html string) []map[string]any {
	rows := parseHTMLTableRows(html)
	if len(rows) == 0 {
		rows = parseTableRowsWithTDHeaders(html, isExamHeaderSet)
	}
	for _, row := range rows {
		normalizeExamRow(row)
	}
	return rows
}

func parseTableRowsWithTDHeaders(html string, acceptHeaders func([]string) bool) []map[string]any {
	tables := extractTags(html, "table")
	for _, table := range tables {
		trs := extractTableRows(table)
		if len(trs) < 2 {
			continue
		}

		headers := extractTagTexts(trs[0], "td")
		if len(headers) == 0 || !acceptHeaders(headers) {
			continue
		}

		out := make([]map[string]any, 0, len(trs)-1)
		for _, row := range trs[1:] {
			cells := extractTagTexts(row, "td")
			if len(cells) == 0 {
				continue
			}

			item := map[string]any{"id": len(out) + 1}
			for i, cell := range cells {
				if i < len(headers) && strings.TrimSpace(headers[i]) != "" {
					item[strings.TrimSpace(headers[i])] = strings.TrimSpace(cell)
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

func isExamHeaderSet(headers []string) bool {
	joined := normalizeCourseText(strings.Join(headers, ""))
	markers := []string{"考试", "考核", "课程", "时间", "地点", "座位"}
	hits := 0
	for _, marker := range markers {
		if strings.Contains(joined, marker) {
			hits++
		}
	}
	return hits >= 2
}

func normalizeExamRow(row map[string]any) {
	if row == nil {
		return
	}
	if v := firstRowString(row, "course_name", "课程名称", "课程", "科目", "教学任务"); v != "" {
		row["course_name"] = v
	}
	if v := firstRowString(row, "exam_time", "考试时间", "考试安排", "时间", "考试日期"); v != "" {
		row["exam_time"] = v
	}
	if v := firstRowString(row, "location", "考试地点", "地点", "教室", "考场"); v != "" {
		row["location"] = v
	}
	if v := firstRowString(row, "seat", "座位号", "座位", "座号"); v != "" {
		row["seat"] = v
	}
	if v := firstRowString(row, "exam_type", "考试类别", "考试类型", "类型", "考核方式"); v != "" {
		row["exam_type"] = v
	}
}

func firstRowString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := row[key]; ok {
			s := strings.TrimSpace(anyToString(v))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

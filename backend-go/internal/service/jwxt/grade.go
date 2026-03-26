package jwxt

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *JwxtDirectService) GetGrade(sess *CachedJWXTSession, semesterID string) (map[string]any, error) {
	if strings.TrimSpace(semesterID) == "" {
		return map[string]any{"success": true, "grades": []any{}, "statistics": map[string]any{}, "message": "请选择学期"}, nil
	}

	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("semesterId", semesterID)
	form.Set("projectType", "")
	form.Set("_", strconv.FormatInt(time.Now().UnixMilli(), 10))

	body, err := s.postForm(client, jwxtBaseURL+"/eams/teach/grade/course/person!search.action", form)
	if err != nil {
		return nil, err
	}

	if strings.Contains(body, "用户名") && strings.Contains(body, "密码") {
		return map[string]any{"success": false, "error": "需要重新登录", "grades": []any{}}, nil
	}

	grades := parseHTMLTableRows(body)
	grades = sanitizeGrades(grades)
	stats := calcGradeStats(grades)
	return map[string]any{
		"success":       true,
		"semester_id":   semesterID,
		"grades":        grades,
		"statistics":    stats,
		"total_courses": len(grades),
	}, nil
}

func calcGradeStats(grades []map[string]any) map[string]any {
	stats := map[string]any{
		"total_courses":      len(grades),
		"average_score":      nil,
		"weighted_average":   nil,
		"total_credits":      0.0,
		"grade_distribution": map[string]int{"90-100": 0, "80-89": 0, "70-79": 0, "60-69": 0, "60以下": 0},
	}
	if len(grades) == 0 {
		return stats
	}

	scores := make([]float64, 0)
	totalCredits := 0.0
	weightedSum := 0.0
	dist := stats["grade_distribution"].(map[string]int)

	for _, g := range grades {
		credits := parseFloatAny(g["学分"])
		totalCredits += credits

		score := -1.0
		fields := []string{"最终成绩", "总评成绩", "成绩", "总评"}
		for _, key := range fields {
			if v, ok := g[key]; ok {
				s := strings.TrimSpace(fmt.Sprintf("%v", v))
				clean := strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), "-", "")
				if clean != "" && isDigits(clean) {
					score = parseFloatAny(v)
					break
				}
			}
		}
		if score >= 0 {
			scores = append(scores, score)
			if credits > 0 {
				weightedSum += score * credits
			}
			switch {
			case score >= 90:
				dist["90-100"]++
			case score >= 80:
				dist["80-89"]++
			case score >= 70:
				dist["70-79"]++
			case score >= 60:
				dist["60-69"]++
			default:
				dist["60以下"]++
			}
		}
	}

	if len(scores) > 0 {
		sum := 0.0
		for _, s := range scores {
			sum += s
		}
		stats["average_score"] = round2(sum / float64(len(scores)))
		if totalCredits > 0 {
			stats["weighted_average"] = round2(weightedSum / totalCredits)
			stats["total_credits"] = round2(totalCredits)
		}
	}

	return stats
}

func sanitizeGrades(rows []map[string]any) []map[string]any {
	if len(rows) == 0 {
		return rows
	}

	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if isGradeHeaderLikeRow(row) {
			continue
		}
		out = append(out, row)
	}
	return out
}

func isGradeHeaderLikeRow(row map[string]any) bool {
	if len(row) == 0 {
		return true
	}

	courseName := strings.TrimSpace(fmt.Sprintf("%v", row["课程名称"]))
	if courseName == "" {
		courseName = strings.TrimSpace(fmt.Sprintf("%v", row["课程"]))
	}

	if courseName != "" {
		n := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", "：", "", ":", "").Replace(courseName)
		if n == "课程名称" || n == "课程名" || n == "课程" || strings.Contains(n, "课程名称") {
			return true
		}
	}

	markerCount := 0
	markers := []string{"课程名称", "课程", "学分", "成绩", "总评", "总评成绩", "最终", "最终成绩", "绩点"}
	for _, key := range markers {
		v, ok := row[key]
		if !ok {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == key {
			markerCount++
		}
	}

	return markerCount >= 2
}

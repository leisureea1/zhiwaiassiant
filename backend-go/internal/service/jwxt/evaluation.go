package jwxt

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (s *JwxtDirectService) GetEvaluationPending(sess *CachedJWXTSession) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}
	body, err := s.get(client, jwxtBaseURL+"/eams/quality/stdEvaluate.action")
	if err != nil {
		return nil, err
	}
	items := parseEvaluationPending(body)
	return map[string]any{"success": true, "total": len(items), "evaluations": items}, nil
}

func (s *JwxtDirectService) AutoEvaluation(sess *CachedJWXTSession) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	pending, _ := s.GetEvaluationPending(sess)
	list, _ := pending["evaluations"].([]map[string]any)
	if len(list) == 0 {
		return map[string]any{
			"success":   true,
			"total":     0,
			"failed":    0,
			"succeeded": 0,
			"message":   "没有待评教课程",
			"details":   []any{},
			"results":   []any{},
		}, nil
	}

	results := make([]map[string]any, 0, len(list))
	details := make([]map[string]any, 0, len(list))
	succeeded := 0
	failed := 0

	for _, item := range list {
		lessonID, _ := item["lesson_id"].(string)
		courseName, _ := item["course_name"].(string)
		teacherName, _ := item["teacher_name"].(string)
		q, err := s.get(client, jwxtBaseURL+"/eams/quality/stdEvaluate!answer.action?evaluationLesson.id="+url.QueryEscape(lessonID))
		if err != nil {
			failed++
			msg := err.Error()
			results = append(results, map[string]any{"evaluationId": lessonID, "success": false, "message": msg})
			details = append(details, map[string]any{
				"lesson_id": lessonID,
				"course":    courseName,
				"teacher":   teacherName,
				"success":   false,
				"message":   msg,
			})
			continue
		}

		payload := buildEvaluationSubmitPayload(q, lessonID)
		form := url.Values{}
		for k, v := range payload {
			form.Set(k, v)
		}
		_, err = s.postForm(client, jwxtBaseURL+"/eams/quality/stdEvaluate!finishAnswer.action", form)
		if err != nil {
			failed++
			msg := err.Error()
			results = append(results, map[string]any{"evaluationId": lessonID, "success": false, "message": msg})
			details = append(details, map[string]any{
				"lesson_id": lessonID,
				"course":    courseName,
				"teacher":   teacherName,
				"success":   false,
				"message":   msg,
			})
			continue
		}
		succeeded++
		results = append(results, map[string]any{"evaluationId": lessonID, "success": true})
		details = append(details, map[string]any{
			"lesson_id": lessonID,
			"course":    courseName,
			"teacher":   teacherName,
			"success":   true,
			"message":   "评教成功",
		})
		time.Sleep(300 * time.Millisecond)
	}

	message := "评教完成"
	if failed > 0 {
		message = fmt.Sprintf("评教完成：成功 %d 门，失败 %d 门", succeeded, failed)
	}

	return map[string]any{
		"success":   true,
		"total":     len(list),
		"failed":    failed,
		"succeeded": succeeded,
		"message":   message,
		"details":   details,
		"results":   results,
	}, nil
}

func parseEvaluationPending(htmlStr string) []map[string]any {
	out := make([]map[string]any, 0)

	// 按 </tr> 分割，找到包含 evalLesson 的行
	chunks := strings.Split(htmlStr, "</tr>")
	lessonRe := regexp.MustCompile(`evaluationLesson\.id=(\d+)`)
	tdRe := regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)

	for _, chunk := range chunks {
		m := lessonRe.FindStringSubmatch(chunk)
		if len(m) < 2 {
			continue
		}
		lessonID := m[1]

		tds := tdRe.FindAllStringSubmatch(chunk, -1)
		if len(tds) < 4 {
			continue
		}

		out = append(out, map[string]any{
			"lesson_id":    lessonID,
			"course_code":  cleanTableCell(tds[0][1]),
			"course_name":  cleanTableCell(tds[1][1]),
			"course_type":  cleanTableCell(tds[2][1]),
			"teacher_name": cleanTableCell(tds[3][1]),
		})
	}
	return out
}

func cleanTableCell(raw string) string {
	s := html.UnescapeString(stripTags(raw))
	s = strings.NewReplacer(
		"\u00a0", " ",
		"\r", " ",
		"\n", " ",
		"\t", " ",
		",", " ",
	).Replace(s)
	// 移除 "(进行评教)"
	s = strings.ReplaceAll(s, "(进行评教)", "")
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

var multiSpaceRe = regexp.MustCompile(`\s{2,}`)

func buildEvaluationSubmitPayload(questionPage string, lessonID string) map[string]string {
	p := map[string]string{
		"evaluationLesson.id": lessonID,
		"semester.id":         "209",
		"teacher.id":          "",
	}
	if sem := regexp.MustCompile(`name="semester\.id"\s+value="(\d+)"`).FindStringSubmatch(questionPage); len(sem) > 1 {
		p["semester.id"] = sem[1]
	}

	qBlock := regexp.MustCompile(`(?is)QUESTIONS\s*=\s*new\s+Questions\(eval\('(\[.*?\])'\)\)`).FindStringSubmatch(questionPage)
	if len(qBlock) < 2 {
		p["result1Num"] = "0"
		p["result2Num"] = "0"
		return p
	}

	jsonStr := strings.ReplaceAll(qBlock[1], `\\`, `\`)
	jsonStr = strings.ReplaceAll(jsonStr, `\"`, `"`)

	var questions []map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &questions); err != nil {
		p["result1Num"] = "0"
		p["result2Num"] = "0"
		return p
	}

	result1 := 0
	result2 := 0
	for _, q := range questions {
		if fmt.Sprintf("%v", q["type"]) == "subtitle" {
			continue
		}
		questionName := fmt.Sprintf("%v", q["name"])
		questionType := fmt.Sprintf("%v", q["questionType"])
		proportion := parseFloatAny(q["proportion"])
		if proportion == 0 {
			proportion = 0.05
		}
		if options, ok := q["options"].([]any); ok && len(options) > 0 {
			opt, _ := options[0].(map[string]any)
			optName := fmt.Sprintf("%v", opt["name"])
			optProportion := parseFloatAny(opt["proportion"])
			if optProportion == 0 {
				optProportion = 1
			}
			score := proportion * optProportion * 100
			p[fmt.Sprintf("result1_%d.questionName", result1)] = questionName
			p[fmt.Sprintf("result1_%d.questionType", result1)] = questionType
			p[fmt.Sprintf("result1_%d.content", result1)] = optName
			p[fmt.Sprintf("result1_%d.score", result1)] = fmt.Sprintf("%.2f", score)
			result1++
		} else {
			p[fmt.Sprintf("result2_%d.questionName", result2)] = questionName
			p[fmt.Sprintf("result2_%d.questionType", result2)] = questionType
			p[fmt.Sprintf("result2_%d.content", result2)] = "良好"
			result2++
		}
	}

	p["result1Num"] = strconv.Itoa(result1)
	p["result2Num"] = strconv.Itoa(result2)
	return p
}

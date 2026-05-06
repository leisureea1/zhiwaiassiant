package jwxt

import (
	"net/url"
	"strings"
)

func (s *JwxtDirectService) GetExam(sess *CachedJWXTSession, semesterID string) (map[string]any, error) {
	client, err := s.clientFromSession(sess)
	if err != nil {
		return nil, err
	}

	examURL := jwxtBaseURL + "/eams/stdExamTable!examTable.action"
	if strings.TrimSpace(semesterID) != "" {
		examURL += "?semester.id=" + url.QueryEscape(semesterID)
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

	exams := parseHTMLTableRows(body)
	return map[string]any{"success": true, "exams": exams, "semester": semesterID}, nil
}

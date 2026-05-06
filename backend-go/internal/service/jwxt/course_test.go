package jwxt

import "testing"

func TestParseCourseActivitiesFiltersOnlineCourses(t *testing.T) {
	html := `
var teachers = [];
activity = new TaskActivity('1','张老师','x','大学英语(网上课程)','x','网上课程','011111111111111111111');
index = 0 * unitCount + 1;
var teachers = [];
activity = new TaskActivity('2','李老师','x','高等数学','x','教学楼A101','011111111111111111111');
index = 1 * unitCount + 2;
`

	courses := parseCourseActivities(html)
	if len(courses) != 1 {
		t.Fatalf("expected 1 offline course, got %d: %#v", len(courses), courses)
	}
	if courses[0]["name"] != "高等数学" {
		t.Fatalf("expected offline course to remain, got %#v", courses[0])
	}
}

func TestIsOnlineCourseRecognizesFullWidthParentheses(t *testing.T) {
	if !isOnlineCourse("大学英语（网络课程）", "大学英语（网络课程）", "") {
		t.Fatal("expected full-width parenthesized online marker to be filtered")
	}
}

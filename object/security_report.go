// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"
	"time"

	"github.com/casdoor/casdoor/edition"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type SecurityPostureCheck struct {
	Key string `json:"key"`
	Ok  bool   `json:"ok"`
}

var securityPostureWeights = map[string]float64{
	"adminMfa": 20, "inactive": 15, "session": 10, "passwordCheck": 12, "stuffing": 12, "signinAlert": 10, "certs": 11, "masking": 10,
}

// GetSecurityPosture runs the checks of the dashboard's security score and weighs them into a score of 0 to 100,
// the same as getSecurityChecks and getPostureScore of web/src/components/dashboard/EditionDashboard.tsx
func GetSecurityPosture(data *DashboardSecurity) ([]*SecurityPostureCheck, int) {
	orgs := data.Organizations
	policy := func(key string) int { return data.Policies[key] }
	missing := func(key string) int { return max(0, orgs-policy(key)) }

	checks := []*SecurityPostureCheck{
		{"adminMfa", data.AdminsWithoutMfa == 0},
		{"inactive", data.InactiveUsers == 0 || (orgs > 0 && missing("disableInactive") == 0)},
		{"session", data.Applications == 0 || data.ApplicationsWithSessionControl == data.Applications || (orgs > 0 && missing("inactiveTimeout") == 0)},
		{"passwordCheck", orgs > 0 && missing("passwordCheck") == 0},
		{"stuffing", orgs > 0 && missing("stuffing") == 0},
		{"signinAlert", orgs > 0 && missing("signinAlert") == 0},
		{"certs", data.CertsExpiring == 0},
		{"masking", orgs > 0 && missing("masking") == 0},
	}

	share := func(check *SecurityPostureCheck) float64 {
		if check.Ok {
			return 1
		}
		switch {
		case check.Key == "adminMfa" && data.Admins > 0:
			return 1 - float64(data.AdminsWithoutMfa)/float64(data.Admins)
		case check.Key == "inactive" && data.Users > 0:
			return math.Max(0, 1-float64(data.InactiveUsers)/float64(data.Users))
		case check.Key == "session" && data.Applications > 0:
			return float64(data.ApplicationsWithSessionControl) / float64(data.Applications)
		case (check.Key == "passwordCheck" || check.Key == "stuffing" || check.Key == "signinAlert" || check.Key == "masking") && orgs > 0:
			return float64(policy(check.Key)) / float64(orgs)
		}
		return 0
	}
	total := 0.0
	for _, check := range checks {
		total += securityPostureWeights[check.Key] * share(check)
	}
	return checks, int(math.Round(total))
}

// renderSigninChart draws the successful and failed sign-ins of each day as stacked bars
func renderSigninChart(days []*SecurityDay) ([]byte, int, int) {
	const width, height, left, right, top, bottom = 1200, 420, 64, 24, 20, 44
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	fill := func(x0, y0, x1, y1 int, c color.Color) {
		draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{C: c}, image.Point{}, draw.Src)
	}
	text := func(x, y int, s string) {
		drawer := &font.Drawer{Dst: img, Src: image.NewUniform(color.RGBA{R: 90, G: 96, B: 110, A: 255}), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
		drawer.DrawString(s)
	}

	maxValue := 1
	for _, day := range days {
		maxValue = max(maxValue, day.Success+day.Failure)
	}
	// a round top for the axis: 1, 2 or 5 times a power of ten
	step := math.Pow(10, math.Floor(math.Log10(float64(maxValue)/4)))
	for _, m := range []float64{1, 2, 5, 10} {
		if float64(maxValue)/(step*m) <= 4 {
			step *= m
			break
		}
	}
	if step < 1 {
		step = 1
	}
	axisMax := step * math.Ceil(float64(maxValue)/step)
	plotHeight := height - top - bottom
	plotWidth := width - left - right
	y := func(value float64) int { return height - bottom - int(value/axisMax*float64(plotHeight)) }

	grid := color.RGBA{R: 229, G: 231, B: 235, A: 255}
	for value := 0.0; value <= axisMax+0.5; value += step {
		fill(left, y(value), width-right, y(value)+1, grid)
		label := fmt.Sprintf("%d", int(value))
		text(left-8-7*len(label), y(value)+4, label)
	}

	if len(days) > 0 {
		slot := float64(plotWidth) / float64(len(days))
		barWidth := max(2, int(slot*0.7))
		labelEvery := max(1, int(math.Ceil(float64(len(days))*42/float64(plotWidth))))
		for i, day := range days {
			x0 := left + int(slot*float64(i)+(slot-float64(barWidth))/2)
			fill(x0, y(float64(day.Success)), x0+barWidth, y(0), color.RGBA{R: 37, G: 99, B: 235, A: 255})
			fill(x0, y(float64(day.Success+day.Failure)), x0+barWidth, y(float64(day.Success)), color.RGBA{R: 220, G: 38, B: 38, A: 255})
			if i%labelEvery == 0 && len(day.Date) == 10 {
				text(x0+barWidth/2-17, height-bottom+18, day.Date[5:])
			}
		}
	}
	fill(left, height-bottom, width-right, height-bottom+1, color.RGBA{R: 156, G: 163, B: 175, A: 255})

	buf := &bytes.Buffer{}
	_ = png.Encode(buf, img)
	return buf.Bytes(), width, height
}

type securityReportText struct {
	lang string
}

func (t securityReportText) pick(zh string, en string) string {
	if strings.HasPrefix(t.lang, "zh") {
		return zh
	}
	return en
}

func (t securityReportText) location(location string) string {
	switch location {
	case "":
		return t.pick("未知", "Unknown")
	case "intranet":
		return t.pick("内网", "Intranet")
	}
	return location
}

func formatChange(current float64, previous float64, isRate bool) string {
	diff := current - previous
	if isRate {
		return fmt.Sprintf("%+.1f pp", diff*100)
	}
	if previous == 0 {
		if current == 0 {
			return "0"
		}
		return fmt.Sprintf("%+d", int(diff))
	}
	return fmt.Sprintf("%+d (%+.0f%%)", int(diff), diff/previous*100)
}

func formatReportTime(value string) string {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return t.In(time.Local).Format("2006-01-02 15:04")
}

// GetSecurityReportDocx writes the security report of an organization for the period as a Word document, in Chinese
// or English after the language, and returns it with its file name
func GetSecurityReportDocx(owner string, from time.Time, to time.Time, isMonth bool, lang string) ([]byte, string, error) {
	t := securityReportText{lang: lang}
	organization, err := getOrganization("admin", owner)
	if err != nil {
		return nil, "", err
	}
	if organization == nil {
		return nil, "", fmt.Errorf("the organization: %s does not exist", owner)
	}
	orgName := organization.DisplayName
	if orgName == "" {
		orgName = organization.Name
	}

	// a period that hasn't ended is counted up to today and compared with as many days before it
	now := time.Now()
	isPartial := false
	if tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, from.Location()); to.After(tomorrow) && from.Before(tomorrow) {
		to, isPartial = tomorrow, true
	}
	prevFrom, prevTo := from.Add(-to.Sub(from)), from
	if isMonth {
		prevFrom = from.AddDate(0, -1, 0)
		if isPartial {
			prevTo = prevFrom.AddDate(0, 0, int(to.Sub(from).Hours()/24+0.5))
			if prevTo.After(from) {
				prevTo = from
			}
		}
	}
	overview, err := GetSecurityOverview(owner, from, to)
	if err != nil {
		return nil, "", err
	}
	previous, err := GetSecurityOverview(owner, prevFrom, prevTo)
	if err != nil {
		return nil, "", err
	}
	security, err := GetDashboardSecurity(owner)
	if err != nil {
		return nil, "", err
	}
	checks, score := GetSecurityPosture(security)
	compliance, err := GetComplianceReport(owner, lang)
	if err != nil {
		return nil, "", err
	}
	orphans, err := GetOrphanAccounts(owner, dashboardInactiveDays)
	if err != nil {
		return nil, "", err
	}
	conflicts, err := GetSodConflicts(owner)
	if err != nil {
		return nil, "", err
	}
	expiringCerts, err := getExpiringCertNames(owner, time.Now(), dashboardCertAlertDays)
	if err != nil {
		return nil, "", err
	}
	mfa, err := GetDashboardMfaCoverage(owner)
	if err != nil {
		return nil, "", err
	}

	lastDay := to.AddDate(0, 0, -1)
	period := fmt.Sprintf("%s ~ %s", from.Format("2006-01-02"), lastDay.Format("2006-01-02"))
	title := t.pick("身份安全报告", "Identity security report")
	fileSuffix := from.Format("20060102") + "-" + lastDay.Format("20060102")
	if isMonth {
		title = t.pick(fmt.Sprintf("%d 年 %d 月身份安全月报", from.Year(), int(from.Month())), "Identity security report for "+from.Format("January 2006"))
		fileSuffix = from.Format("2006-01")
	}
	fileName := fmt.Sprintf("security-report-%s-%s.docx", owner, fileSuffix)

	d := &docxBuilder{}
	d.Title(title)
	d.paragraph("", "center", docxRun(orgName, true, "2E5597", 32))
	d.paragraph("Spacer", "", "")

	license := edition.GetLicenseInfo()
	licensee := "-"
	if license.Licensee != "" {
		licensee = license.Licensee
	}
	cover := [][]docxCell{
		{{Text: t.pick("组织", "Organization"), Bold: true}, {Text: fmt.Sprintf("%s (%s)", orgName, owner)}},
		{{Text: t.pick("统计周期", "Period"), Bold: true}, {Text: period}},
		{{Text: t.pick("对比周期", "Compared with"), Bold: true}, {Text: fmt.Sprintf("%s ~ %s", prevFrom.Format("2006-01-02"), prevTo.AddDate(0, 0, -1).Format("2006-01-02"))}},
		{{Text: t.pick("生成时间", "Generated"), Bold: true}, {Text: now.Format("2006-01-02 15:04:05 (UTC-07:00)")}},
		{{Text: t.pick("授权单位", "Licensed to"), Bold: true}, {Text: licensee}},
		{{Text: t.pick("产品版本", "Edition"), Bold: true}, {Text: "Casdoor " + license.Edition}},
	}
	d.Table(nil, []int{28, 72}, cover)
	d.Note(t.pick("本报告由 Casdoor 根据该组织的审计日志、登录记录和配置自动生成，数字以生成时数据库中的记录为准。",
		"Casdoor generated this report from the organization's audit records, sign-in traces and settings, the numbers are what the database held when it was generated."))
	if isPartial {
		d.Note(t.pick("统计周期尚未结束，只统计到生成当天，并与上一周期相同天数对比。",
			"The period has not ended yet, it is counted up to the day of the report and compared with as many days of the previous period."))
	}
	if overview.Truncated || previous.Truncated {
		d.Note(t.pick(fmt.Sprintf("记录过多，只统计了每个周期的前 %d 条记录。", securityOverviewMaxRecords),
			fmt.Sprintf("There were too many records, only the first %d of each period were counted.", securityOverviewMaxRecords)))
	}
	d.PageBreak()

	// 1. summary
	d.Heading(1, t.pick("一、总体情况", "1. Summary"))
	d.Paragraph(t.pick(
		fmt.Sprintf("本期共登录成功 %d 次、失败 %d 次，失败率 %.1f%%（上期 %.1f%%）；%d 个用户登录过，新注册 %d 人；管理员操作 %d 次，被拒绝的请求 %d 次；出现新登录地点 %d 个、新设备 %d 台。",
			overview.SigninSuccess, overview.SigninFailure, overview.FailureRate*100, previous.FailureRate*100, overview.ActiveUsers, overview.Signups, overview.AdminOperations, overview.Denied, overview.NewLocations, overview.NewDevices),
		fmt.Sprintf("In this period there were %d successful and %d failed sign-ins, a failure rate of %.1f%% (%.1f%% in the previous period). %d users signed in and %d signed up. Admins made %d changes, %d requests were denied, and users signed in from %d new places and %d new devices.",
			overview.SigninSuccess, overview.SigninFailure, overview.FailureRate*100, previous.FailureRate*100, overview.ActiveUsers, overview.Signups, overview.AdminOperations, overview.Denied, overview.NewLocations, overview.NewDevices)))

	metric := func(name string, current int, prev int) []docxCell {
		return []docxCell{{Text: name}, {Text: fmt.Sprintf("%d", current)}, {Text: fmt.Sprintf("%d", prev)}, {Text: formatChange(float64(current), float64(prev), false)}}
	}
	d.Table([]string{t.pick("指标", "Metric"), t.pick("本期", "This period"), t.pick("上期", "Previous period"), t.pick("变化", "Change")}, []int{40, 20, 20, 20}, [][]docxCell{
		metric(t.pick("登录成功", "Successful sign-ins"), overview.SigninSuccess, previous.SigninSuccess),
		metric(t.pick("登录失败", "Failed sign-ins"), overview.SigninFailure, previous.SigninFailure),
		{{Text: t.pick("失败率", "Failure rate")}, {Text: fmt.Sprintf("%.1f%%", overview.FailureRate*100)}, {Text: fmt.Sprintf("%.1f%%", previous.FailureRate*100)}, {Text: formatChange(overview.FailureRate, previous.FailureRate, true)}},
		metric(t.pick("活跃用户", "Active users"), overview.ActiveUsers, previous.ActiveUsers),
		metric(t.pick("新注册用户", "Sign-ups"), overview.Signups, previous.Signups),
		metric(t.pick("管理员操作", "Admin operations"), overview.AdminOperations, previous.AdminOperations),
		metric(t.pick("被拒绝的请求", "Denied requests"), overview.Denied, previous.Denied),
		metric(t.pick("新登录地点", "New locations"), overview.NewLocations, previous.NewLocations),
		metric(t.pick("新登录设备", "New devices"), overview.NewDevices, previous.NewDevices),
	})

	grade := t.pick("较好", "Strong")
	gradeColor := docxColorPass
	if score < 65 {
		grade, gradeColor = t.pick("需关注", "Needs attention"), docxColorFail
	} else if score < 85 {
		grade, gradeColor = t.pick("一般", "Fair"), docxColorWarn
	}
	d.Heading(2, t.pick("安全评分", "Security score"))
	d.Rich(docxCell{Text: t.pick("当前安全评分：", "Current security score: ")}, docxCell{Text: fmt.Sprintf("%d / 100（%s）", score, grade), Bold: true, Color: gradeColor})
	checkNames := map[string][2]string{
		"adminMfa":      {"管理员开启多因素认证", "MFA for admins"},
		"inactive":      {"长期未用的账号", "Inactive accounts"},
		"session":       {"会话控制（空闲超时、互斥登录）", "Session control (idle timeout, exclusive sign-in)"},
		"passwordCheck": {"弱口令和泄露口令检查", "Weak and leaked password checks"},
		"stuffing":      {"撞库防护", "Credential stuffing defense"},
		"signinAlert":   {"异地和新设备登录提醒", "Alerts on new places and devices"},
		"certs":         {"签名密钥和证书有效期", "Signing keys and certificates"},
		"masking":       {"个人信息脱敏", "Personal data masking"},
	}
	checkRows := [][]docxCell{}
	for _, check := range checks {
		name := checkNames[check.Key]
		status := docxCell{Text: t.pick("达标", "Met"), Color: docxColorPass, Bold: true}
		if !check.Ok {
			status = docxCell{Text: t.pick("未达标", "Not met"), Color: docxColorFail, Bold: true}
		}
		checkRows = append(checkRows, []docxCell{{Text: t.pick(name[0], name[1])}, {Text: fmt.Sprintf("%.0f", securityPostureWeights[check.Key])}, status})
	}
	d.Table([]string{t.pick("检查项", "Check"), t.pick("权重", "Weight"), t.pick("结果", "Result")}, []int{60, 15, 25}, checkRows)
	mfaRate := 0.0
	if mfa.Total > 0 {
		mfaRate = float64(mfa.Enabled) / float64(mfa.Total) * 100
	}
	d.Paragraph(t.pick(
		fmt.Sprintf("用户 %d 人，其中 %d 人开启了多因素认证（%.1f%%）；管理员 %d 人，未开启多因素认证的 %d 人。",
			mfa.Total, mfa.Enabled, mfaRate, security.Admins, security.AdminsWithoutMfa),
		fmt.Sprintf("%d users, %d of them with MFA (%.1f%%); %d admins, %d of them without MFA.",
			mfa.Total, mfa.Enabled, mfaRate, security.Admins, security.AdminsWithoutMfa)))

	// 2. trends
	d.Heading(1, t.pick("二、登录趋势", "2. Sign-in trend"))
	chart, width, height := renderSigninChart(overview.Days)
	d.Image(chart, width, height, 16.5)
	d.Note(t.pick("图 1  每日登录次数（蓝色为成功，红色为失败）", "Figure 1  Sign-ins per day (blue: successful, red: failed)"))
	dayRows := [][]docxCell{}
	for _, day := range overview.Days {
		failure := docxCell{Text: fmt.Sprintf("%d", day.Failure)}
		if day.Failure > 0 && day.Failure*5 >= day.Success+day.Failure && day.Failure >= 5 {
			failure.Color, failure.Bold = docxColorFail, true
		}
		dayRows = append(dayRows, []docxCell{{Text: day.Date}, {Text: fmt.Sprintf("%d", day.Success)}, failure, {Text: fmt.Sprintf("%d", day.Signups)}})
	}
	d.Table([]string{t.pick("日期", "Date"), t.pick("登录成功", "Successful"), t.pick("登录失败", "Failed"), t.pick("新注册", "Sign-ups")}, []int{34, 22, 22, 22}, dayRows)

	// 3. sign-in risks
	d.Heading(1, t.pick("三、登录风险", "3. Sign-in risks"))
	d.Heading(2, t.pick("登录失败最多的 IP", "IPs with the most failed sign-ins"))
	traces := []*SigninTrace{}
	err = ormer.Engine.Where("owner = ? and kind = ?", owner, "location").Find(&traces)
	if err != nil {
		return nil, "", err
	}
	locator := newIpLocator(traces)
	ipRows := [][]docxCell{}
	for _, item := range overview.TopFailedIps {
		ipRows = append(ipRows, []docxCell{{Text: item.Name}, {Text: t.location(locator(item.Name))}, {Text: fmt.Sprintf("%d", item.Count)}})
	}
	if len(ipRows) == 0 {
		d.Paragraph(t.pick("本期没有登录失败。", "No failed sign-in in this period."))
	} else {
		d.Table([]string{"IP", t.pick("位置", "Location"), t.pick("失败次数", "Failures")}, []int{35, 45, 20}, ipRows)
	}
	d.Heading(2, t.pick("登录失败最多的账号", "Accounts with the most failed sign-ins"))
	userRows := [][]docxCell{}
	for _, item := range overview.TopFailedUsers {
		userRows = append(userRows, []docxCell{{Text: item.Name}, {Text: fmt.Sprintf("%d", item.Count)}})
	}
	if len(userRows) == 0 {
		d.Paragraph(t.pick("本期没有登录失败。", "No failed sign-in in this period."))
	} else {
		d.Table([]string{t.pick("账号", "Account"), t.pick("失败次数", "Failures")}, []int{75, 25}, userRows)
	}
	d.Heading(2, t.pick("新登录地点和设备", "New places and devices"))
	if len(overview.RecentSigninRisks) == 0 {
		d.Paragraph(t.pick("本期没有从新地点或新设备登录。若组织未开启“异地和新设备登录提醒”，这里不会有记录。",
			"Nobody signed in from a new place or device in this period. Nothing is recorded unless the organization turns on the sign-in alerts."))
	} else {
		riskRows := [][]docxCell{}
		for _, trace := range overview.RecentSigninRisks {
			kind, detail := t.pick("地点", "Location"), t.location(trace.TraceKey)
			if trace.Kind != "location" {
				kind, detail = t.pick("设备", "Device"), trace.UserAgent
			}
			riskRows = append(riskRows, []docxCell{{Text: formatReportTime(trace.CreatedTime)}, {Text: trace.UserName}, {Text: kind}, {Text: detail}, {Text: trace.Ip}})
		}
		d.Table([]string{t.pick("时间", "Time"), t.pick("用户", "User"), t.pick("类型", "Type"), t.pick("地点或设备", "Place or device"), "IP"}, []int{19, 15, 10, 38, 18}, riskRows)
	}

	// 4. admin operations
	d.Heading(1, t.pick("四、管理员操作", "4. Admin operations"))
	d.Paragraph(t.pick(fmt.Sprintf("本期管理员共做了 %d 次修改操作（不含登录、注册等普通用户操作），被拒绝的请求 %d 次。", overview.AdminOperations, overview.Denied),
		fmt.Sprintf("Admins made %d changes in this period (sign-ins, sign-ups and other end-user actions excluded), and %d requests were denied.", overview.AdminOperations, overview.Denied)))
	countTable := func(nameHeader string, items []*SecurityCount) {
		rows := [][]docxCell{}
		for _, item := range items {
			rows = append(rows, []docxCell{{Text: item.Name}, {Text: fmt.Sprintf("%d", item.Count)}})
		}
		if len(rows) == 0 {
			d.Paragraph(t.pick("无", "None"))
			return
		}
		d.Table([]string{nameHeader, t.pick("次数", "Count")}, []int{75, 25}, rows)
	}
	d.Heading(2, t.pick("最常见的操作", "Most frequent operations"))
	countTable(t.pick("操作（接口）", "Operation (API)"), overview.TopActions)
	d.Heading(2, t.pick("操作最多的管理员", "Most active admins"))
	countTable(t.pick("管理员", "Admin"), overview.TopAdmins)

	// 5. MLPS
	d.Heading(1, t.pick("五、等保 2.0 三级与密评自查", "5. MLPS 2.0 level 3 and cryptography self-check"))
	d.Rich(docxCell{Text: t.pick("共检查 ", "Checked ")}, docxCell{Text: fmt.Sprintf("%d", len(compliance.Items)), Bold: true},
		docxCell{Text: t.pick(" 项：符合 ", " items: ")}, docxCell{Text: fmt.Sprintf("%d", compliance.Pass), Bold: true, Color: docxColorPass},
		docxCell{Text: t.pick(" 项，部分符合 ", " passed, ")}, docxCell{Text: fmt.Sprintf("%d", compliance.Warn), Bold: true, Color: docxColorWarn},
		docxCell{Text: t.pick(" 项，不符合 ", " warnings, ")}, docxCell{Text: fmt.Sprintf("%d", compliance.Fail), Bold: true, Color: docxColorFail},
		docxCell{Text: t.pick(" 项。", " failed.")})
	statusCell := func(status string) docxCell {
		switch status {
		case ComplianceStatusPass:
			return docxCell{Text: t.pick("符合", "Pass"), Color: docxColorPass, Bold: true}
		case ComplianceStatusWarn:
			return docxCell{Text: t.pick("部分符合", "Warning"), Color: docxColorWarn, Bold: true}
		}
		return docxCell{Text: t.pick("不符合", "Fail"), Color: docxColorFail, Bold: true}
	}
	complianceRows := [][]docxCell{}
	for _, item := range compliance.Items {
		complianceRows = append(complianceRows, []docxCell{{Text: item.Category}, {Text: item.Title}, statusCell(item.Status), {Text: item.Detail}})
	}
	d.Table([]string{t.pick("类别", "Category"), t.pick("检查项", "Item"), t.pick("结果", "Result"), t.pick("说明", "Detail")}, []int{18, 27, 13, 42}, complianceRows)

	// 6. governance
	reasonCounts := map[string]int{}
	for _, account := range orphans {
		for _, reason := range account.Reasons {
			reasonCounts[reason]++
		}
	}
	d.Heading(1, t.pick("六、账号治理", "6. Account governance"))
	d.Table([]string{t.pick("项目", "Item"), t.pick("数量", "Count")}, []int{75, 25}, [][]docxCell{
		{{Text: t.pick(fmt.Sprintf("孤儿账号（%d 天未登录或停用后仍有角色）", dashboardInactiveDays), fmt.Sprintf("Orphan accounts (unused for %d days, or disabled but holding roles)", dashboardInactiveDays))}, {Text: fmt.Sprintf("%d", len(orphans)), Bold: len(orphans) > 0}},
		{{Text: t.pick("　其中从未登录", "  never signed in")}, {Text: fmt.Sprintf("%d", reasonCounts["never-signed-in"])}},
		{{Text: t.pick("　其中长期未登录", "  inactive")}, {Text: fmt.Sprintf("%d", reasonCounts["inactive"])}},
		{{Text: t.pick("　其中已停用仍有角色", "  disabled but holding roles")}, {Text: fmt.Sprintf("%d", reasonCounts["disabled-with-roles"])}},
		{{Text: t.pick("职责分离冲突", "Separation-of-duty conflicts")}, {Text: fmt.Sprintf("%d", len(conflicts)), Bold: len(conflicts) > 0}},
		{{Text: t.pick(fmt.Sprintf("%d 天内到期的证书", dashboardCertAlertDays), fmt.Sprintf("Certificates expiring within %d days", dashboardCertAlertDays))}, {Text: fmt.Sprintf("%d", len(expiringCerts)), Bold: len(expiringCerts) > 0}},
	})
	if len(conflicts) > 0 {
		rows := [][]docxCell{}
		for i, conflict := range conflicts {
			if i >= 20 {
				break
			}
			rows = append(rows, []docxCell{{Text: conflict.User}, {Text: conflict.Rule}, {Text: strings.Join(conflict.Roles, ", ")}})
		}
		d.Table([]string{t.pick("用户", "User"), t.pick("规则", "Rule"), t.pick("角色", "Roles")}, []int{30, 30, 40}, rows)
	}

	// 7. AI agents
	d.Heading(1, t.pick("七、AI 智能体与 MCP 工具调用", "7. AI agents and MCP tool calls"))
	d.Paragraph(t.pick(
		fmt.Sprintf("本期智能体和用户经 MCP 网关调用工具 %d 次（上期 %d 次），涉及智能体 %d 个；被拦截 %d 次（上期 %d 次），等待人工审批 %d 次。",
			overview.McpCalls, previous.McpCalls, overview.McpAgents, overview.McpBlocked, previous.McpBlocked, overview.McpPendingApproval),
		fmt.Sprintf("Agents and users called tools through the MCP gateway %d times in this period (%d in the previous period), by %d agents; %d calls were blocked (%d in the previous period) and %d waited for approval.",
			overview.McpCalls, previous.McpCalls, overview.McpAgents, overview.McpBlocked, previous.McpBlocked, overview.McpPendingApproval)))
	d.Heading(2, t.pick("调用最多的工具", "Most called tools"))
	countTable(t.pick("MCP 服务器 / 工具", "MCP server / tool"), overview.TopMcpTools)
	d.Heading(2, t.pick("调用最多的智能体", "Most active agents"))
	countTable(t.pick("智能体", "Agent"), overview.TopMcpAgents)
	d.Heading(2, t.pick("被拦截最多的工具", "Most blocked tools"))
	countTable(t.pick("MCP 服务器 / 工具", "MCP server / tool"), overview.TopMcpBlockedTools)

	// 8. findings
	d.Heading(1, t.pick("八、发现与建议", "8. Findings and recommendations"))
	findings := getSecurityReportFindings(t, overview, previous, security, checks, compliance, len(orphans), len(conflicts), expiringCerts)
	for _, finding := range findings {
		d.Bullet(finding)
	}

	footer := fmt.Sprintf("%s  %s", orgName, title)
	data, err := d.Bytes(title, footer, t.pick("第 %s 页  共 %s 页", "Page %s of %s"))
	return data, fileName, err
}

func getSecurityReportFindings(t securityReportText, overview *SecurityOverview, previous *SecurityOverview, security *DashboardSecurity, checks []*SecurityPostureCheck,
	compliance *ComplianceReport, orphans int, conflicts int, expiringCerts []string,
) []string {
	res := []string{}
	add := func(zh string, en string) {
		res = append(res, t.pick(zh, en))
	}

	if overview.SigninFailure >= 10 && overview.FailureRate >= 0.2 {
		add(fmt.Sprintf("登录失败率偏高（%.1f%%），可能存在口令猜测或撞库，建议开启撞库防护和多因素认证，并核查失败最多的 IP 和账号。", overview.FailureRate*100),
			fmt.Sprintf("The failure rate of sign-ins is high (%.1f%%), which suggests password guessing or credential stuffing: turn on the credential stuffing defense and MFA, and look into the IPs and accounts that fail most.", overview.FailureRate*100))
	}
	if previous.SigninFailure > 0 && overview.SigninFailure >= 10 && overview.SigninFailure >= previous.SigninFailure*3/2 {
		add(fmt.Sprintf("登录失败次数比上期增加 %d%%，请留意是否有针对性的攻击。", (overview.SigninFailure-previous.SigninFailure)*100/previous.SigninFailure),
			fmt.Sprintf("Failed sign-ins grew %d%% over the previous period, watch for a targeted attack.", (overview.SigninFailure-previous.SigninFailure)*100/previous.SigninFailure))
	}
	if len(overview.TopFailedIps) > 0 && overview.TopFailedIps[0].Count >= 20 {
		ip := overview.TopFailedIps[0]
		add(fmt.Sprintf("IP %s 本期登录失败 %d 次，建议核实来源，必要时在网关或防火墙中封禁。", ip.Name, ip.Count),
			fmt.Sprintf("The IP %s failed to sign in %d times, find out who it is and block it at the gateway or firewall if needed.", ip.Name, ip.Count))
	}
	if len(overview.TopFailedUsers) > 0 && overview.TopFailedUsers[0].Count >= 10 {
		user := overview.TopFailedUsers[0]
		add(fmt.Sprintf("账号 %s 本期登录失败 %d 次，建议联系本人确认，必要时重置口令。", user.Name, user.Count),
			fmt.Sprintf("The account %s failed to sign in %d times, check with its owner and reset the password if needed.", user.Name, user.Count))
	}
	if security.AdminsWithoutMfa > 0 {
		add(fmt.Sprintf("有 %d 个管理员没有开启多因素认证，管理员账号一旦被盗影响最大，应当强制开启。", security.AdminsWithoutMfa),
			fmt.Sprintf("%d admins have no MFA. A stolen admin account does the most harm, require MFA for them.", security.AdminsWithoutMfa))
	}
	if overview.NewLocations+overview.NewDevices > 0 {
		add(fmt.Sprintf("本期出现新登录地点 %d 个、新设备 %d 台，建议抽查“新登录地点和设备”中的记录是否为本人。", overview.NewLocations, overview.NewDevices),
			fmt.Sprintf("Users signed in from %d new places and %d new devices, spot-check the list of new places and devices.", overview.NewLocations, overview.NewDevices))
	}
	if overview.Denied > 0 {
		add(fmt.Sprintf("有 %d 次请求因权限不足被拒绝，若集中在少数账号，可能是越权尝试。", overview.Denied),
			fmt.Sprintf("%d requests were denied for lack of permission, if they come from a few accounts they may be privilege escalation attempts.", overview.Denied))
	}
	if orphans > 0 {
		add(fmt.Sprintf("有 %d 个孤儿账号，建议在“合规 > 账号治理”中停用。", orphans),
			fmt.Sprintf("There are %d orphan accounts, disable them in Compliance > Account governance.", orphans))
	}
	if conflicts > 0 {
		add(fmt.Sprintf("有 %d 个职责分离冲突，同一用户同时持有互斥角色，应当解除其中一个。", conflicts),
			fmt.Sprintf("There are %d separation-of-duty conflicts where a user holds roles that should be kept apart, remove one of them.", conflicts))
	}
	if len(expiringCerts) > 0 {
		add(fmt.Sprintf("证书 %s 将在 %d 天内到期，请及时轮换。", strings.Join(expiringCerts, "、"), dashboardCertAlertDays),
			fmt.Sprintf("The certificates %s expire within %d days, rotate them.", strings.Join(expiringCerts, ", "), dashboardCertAlertDays))
	}
	if compliance.Fail > 0 || compliance.Warn > 0 {
		add(fmt.Sprintf("等保与密评自查中有 %d 项不符合、%d 项部分符合，见第五节，建议在测评前整改。", compliance.Fail, compliance.Warn),
			fmt.Sprintf("The MLPS and cryptography self-check has %d failed and %d partly met items (section 5), fix them before the assessment.", compliance.Fail, compliance.Warn))
	}
	tips := map[string][2]string{
		"passwordCheck": {"未开启弱口令和泄露口令检查，建议在组织的安全设置中开启。", "Weak and leaked password checks are off, turn them on in the organization's security settings."},
		"stuffing":      {"未开启撞库防护，建议设置同一 IP 短时间内失败账号数的阈值。", "The credential stuffing defense is off, set how many accounts may fail from one IP in a short time."},
		"signinAlert":   {"未开启异地和新设备登录提醒，开启后用户能及时发现账号被盗用。", "Sign-in alerts are off, with them users notice a stolen account early."},
		"masking":       {"未开启个人信息脱敏，控制台中手机号、邮箱、证件号会完整显示。", "Personal data masking is off, the console shows full phone numbers, emails and ID numbers."},
		"session":       {"部分应用没有会话控制，建议设置空闲超时或互斥登录。", "Some applications have no session control, set an idle timeout or exclusive sign-in."},
	}
	for _, check := range checks {
		if tip, ok := tips[check.Key]; ok && !check.Ok {
			add(tip[0], tip[1])
		}
	}
	if len(res) == 0 {
		add("本期未发现明显的安全问题，请保持现有配置并继续关注。", "No notable security issue was found in this period, keep the current settings and keep watching.")
	}
	return res
}

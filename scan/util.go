package scan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/object"
)

func getSiteBaseURLs(site *object.Site) []string {
	res := []string{}
	seen := map[string]struct{}{}

	schemes := []string{"http"}
	lowerSslMode := strings.ToLower(strings.TrimSpace(site.SslMode))
	if strings.Contains(lowerSslMode, "https") || strings.Contains(lowerSslMode, "enable") {
		schemes = []string{"https"}
	} else if lowerSslMode == "" || strings.Contains(lowerSslMode, "auto") {
		schemes = []string{"https", "http"}
	}

	appendCandidate := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}

		if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
			u, err := url.Parse(raw)
			if err != nil || u.Host == "" {
				return
			}
			baseURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			if _, ok := seen[baseURL]; !ok {
				seen[baseURL] = struct{}{}
				res = append(res, baseURL)
			}
			return
		}

		raw = strings.TrimPrefix(raw, "//")
		for _, scheme := range schemes {
			baseURL := fmt.Sprintf("%s://%s", scheme, raw)
			u, err := url.Parse(baseURL)
			if err != nil || u.Host == "" {
				continue
			}
			if _, ok := seen[baseURL]; ok {
				continue
			}
			seen[baseURL] = struct{}{}
			res = append(res, baseURL)
		}
	}

	appendCandidate(site.Domain)
	for _, domain := range site.OtherDomains {
		appendCandidate(domain)
	}
	appendCandidate(site.Host)
	for _, host := range site.Hosts {
		appendCandidate(host)
	}

	if len(res) == 0 {
		appendCandidate(site.GetHost())
	}

	return res
}

func splitScanTargets(raw string) []string {
	parts := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	res := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		res = append(res, part)
	}
	return res
}

func normalizeScanBaseURL(rawURL string) (string, string, error) {
	candidate := strings.TrimSpace(rawURL)
	if candidate == "" {
		return "", "", fmt.Errorf("target URL is empty")
	}

	lower := strings.ToLower(candidate)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		candidate = fmt.Sprintf("https://%s", candidate)
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return "", "", err
	}

	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	if strings.TrimSpace(u.Host) == "" {
		return "", "", fmt.Errorf("invalid target URL: host is empty")
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, u.Host)
	return rawURL, baseURL, nil
}

func getOnlineScanLists(link string) ([]CVE, []Fingerprint, error) {
	request, err := http.NewRequest(http.MethodGet, strings.TrimSpace(link), nil)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, nil, err
	}

	res := onlineScanLists{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, nil, err
	}

	return res.CVEList, res.FingerprintList, nil
}

func mergeOnlineCVEs(target map[string][]CVE, online []CVE) {
	for _, cve := range online {
		name := strings.TrimSpace(cve.Name)
		if name == "" {
			continue
		}

		seen := map[string]struct{}{}
		for _, existingCve := range target[name] {
			if existingCve.Code == "" {
				continue
			}
			seen[existingCve.Code] = struct{}{}
		}

		if cve.Code != "" {
			if _, ok := seen[cve.Code]; ok {
				continue
			}
		}

		target[name] = append(target[name], cve)
	}
}

func mergeOnlineFingerprints(base []Fingerprint, online []Fingerprint) []Fingerprint {
	seen := map[string]struct{}{}
	for _, fingerprint := range base {
		name := strings.TrimSpace(fingerprint.Name)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
	}

	for _, fingerprint := range online {
		name := strings.TrimSpace(fingerprint.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		base = append(base, fingerprint)
	}

	return base
}

func isFingerprintMatched(client *http.Client, baseURL string, fingerprint Fingerprint) (bool, error) {
	method := strings.TrimSpace(fingerprint.HttpInfo.Method)
	if method == "" {
		method = http.MethodGet
	}

	_, headers, body, err := doRequest(client, method, baseURL, fingerprint.HttpInfo.Path)
	if err != nil {
		return false, err
	}

	headersText := buildHeadersText(headers)
	for _, matcher := range fingerprint.HttpInfo.Matchers {
		if isMatcherMatched(client, baseURL, matcher, body, headersText) {
			return true, nil
		}
	}

	return false, nil
}

func isMatcherMatched(client *http.Client, baseURL string, matcher FingerprintMatcher, body string, headersText string) bool {
	value := strings.TrimSpace(matcher.Value)
	if value == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(matcher.Pos)) {
	case "body":
		return strings.Contains(body, value)
	case "header", "headers":
		return strings.Contains(strings.ToLower(headersText), strings.ToLower(value))
	case "icon":
		statusCode, _, iconBody, err := doRequest(client, http.MethodGet, baseURL, "/favicon.ico")
		if err != nil {
			return false
		}
		hashes := []string{
			strconv.FormatUint(uint64(len(iconBody)), 10),
			strconv.FormatUint(uint64(statusCode), 10),
		}
		if strings.Contains(iconBody, value) {
			return true
		}
		for _, hash := range hashes {
			if hash == value {
				return true
			}
		}
	}

	return false
}

func getFingerprintVersion(client *http.Client, baseURL string, fingerprint Fingerprint) (string, error) {
	regex := strings.TrimSpace(fingerprint.VersionInfo.Regex)
	if regex == "" {
		return "", nil
	}

	method := strings.TrimSpace(fingerprint.VersionInfo.Method)
	if method == "" {
		method = http.MethodGet
	}

	_, headers, body, err := doRequest(client, method, baseURL, fingerprint.VersionInfo.Path)
	if err != nil {
		return "", err
	}

	target := body
	if strings.EqualFold(strings.TrimSpace(fingerprint.VersionInfo.Part), "header") {
		target = strings.ToLower(buildHeadersText(headers))
	}

	target = strings.ReplaceAll(target, " ", "")

	re, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(target)
	if len(matches) < 2 {
		return "", nil
	}

	return strings.TrimSpace(matches[1]), nil
}

func doRequest(client *http.Client, method string, baseURL string, path string) (int, http.Header, string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return 0, nil, "", err
	}

	relative, err := url.Parse(strings.TrimSpace(path))
	if err != nil {
		return 0, nil, "", err
	}

	if relative.Path == "" {
		relative.Path = "/"
	}

	requestURL := base.ResolveReference(relative)
	requestMethod := method
	maxRedirects := 8

	for i := 0; i <= maxRedirects; i++ {
		req, err := http.NewRequest(requestMethod, requestURL.String(), nil)
		if err != nil {
			return 0, nil, "", err
		}

		resp, err := client.Do(req)
		if err != nil {
			return 0, nil, "", err
		}

		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		_ = resp.Body.Close()
		if readErr != nil {
			return 0, nil, "", readErr
		}

		if (resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusPermanentRedirect) && i < maxRedirects {
			location := strings.TrimSpace(resp.Header.Get("Location"))
			if location != "" {
				locationURL, parseErr := url.Parse(location)
				if parseErr == nil {
					requestURL = requestURL.ResolveReference(locationURL)
					if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther {
						requestMethod = http.MethodGet
					}
					continue
				}
			}
		}

		return resp.StatusCode, resp.Header.Clone(), string(bodyBytes), nil
	}

	return 0, nil, "", fmt.Errorf("too many redirects")
}

func buildHeadersText(headers http.Header) string {
	if headers == nil {
		return ""
	}

	var builder strings.Builder
	for key, values := range headers {
		for _, value := range values {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(value)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func filterMatchedCVEs(cves []CVE, version string) []CVE {
	matched := make([]CVE, 0, len(cves))
	for _, cve := range cves {
		rule := strings.TrimSpace(cve.Rule)
		if rule == "" {
			matched = append(matched, cve)
			continue
		}

		if strings.TrimSpace(version) == "" {
			continue
		}

		ok, err := evalVersionRule(rule, version)
		if err != nil {
			logs.Warning("scan: failed to evaluate cve rule, cve = %s, rule = %s, version = %s, err = %v", cve.Code, cve.Rule, version, err)
			continue
		}
		if ok {
			matched = append(matched, cve)
		}
	}

	return matched
}

func evalVersionRule(rule string, version string) (bool, error) {
	comparisonRegexp := regexp.MustCompile(`version\s*(<=|>=|==|!=|<|>)\s*"([^"&|)]*)"?`)
	replaced := comparisonRegexp.ReplaceAllStringFunc(rule, func(part string) string {
		match := comparisonRegexp.FindStringSubmatch(part)
		if len(match) != 3 {
			return "false"
		}

		cmp := compareVersion(version, strings.TrimSpace(match[2]))
		switch match[1] {
		case "<":
			return strconv.FormatBool(cmp < 0)
		case "<=":
			return strconv.FormatBool(cmp <= 0)
		case ">":
			return strconv.FormatBool(cmp > 0)
		case ">=":
			return strconv.FormatBool(cmp >= 0)
		case "==":
			return strconv.FormatBool(cmp == 0)
		case "!=":
			return strconv.FormatBool(cmp != 0)
		default:
			return "false"
		}
	})

	if strings.Contains(replaced, "version") {
		return false, fmt.Errorf("invalid version expression: %s", rule)
	}

	return evalBoolExpr(replaced)
}

func evalBoolExpr(expr string) (bool, error) {
	tokens := tokenizeBoolExpr(expr)
	index := 0

	var parseExpr func() (bool, error)
	var parseTerm func() (bool, error)
	var parseFactor func() (bool, error)

	parseExpr = func() (bool, error) {
		left, err := parseTerm()
		if err != nil {
			return false, err
		}
		for index < len(tokens) && tokens[index] == "||" {
			index++
			right, err := parseTerm()
			if err != nil {
				return false, err
			}
			left = left || right
		}
		return left, nil
	}

	parseTerm = func() (bool, error) {
		left, err := parseFactor()
		if err != nil {
			return false, err
		}
		for index < len(tokens) && tokens[index] == "&&" {
			index++
			right, err := parseFactor()
			if err != nil {
				return false, err
			}
			left = left && right
		}
		return left, nil
	}

	parseFactor = func() (bool, error) {
		if index >= len(tokens) {
			return false, fmt.Errorf("unexpected end of expression")
		}

		token := tokens[index]
		index++

		switch token {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "(":
			value, err := parseExpr()
			if err != nil {
				return false, err
			}
			if index >= len(tokens) || tokens[index] != ")" {
				return false, fmt.Errorf("missing closing parenthesis")
			}
			index++
			return value, nil
		default:
			return false, fmt.Errorf("invalid token: %s", token)
		}
	}

	value, err := parseExpr()
	if err != nil {
		return false, err
	}
	if index != len(tokens) {
		return false, fmt.Errorf("invalid expression: %s", expr)
	}

	return value, nil
}

func tokenizeBoolExpr(expr string) []string {
	expr = strings.ReplaceAll(expr, "(", " ( ")
	expr = strings.ReplaceAll(expr, ")", " ) ")
	parts := strings.Fields(expr)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(strings.ToLower(part))
		if token == "" {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func compareVersion(left string, right string) int {
	leftTokens := tokenizeVersion(left)
	rightTokens := tokenizeVersion(right)

	maxLength := len(leftTokens)
	if len(rightTokens) > maxLength {
		maxLength = len(rightTokens)
	}

	for i := 0; i < maxLength; i++ {
		if i >= len(leftTokens) {
			return compareVersionTail(nil, rightTokens[i:])
		}
		if i >= len(rightTokens) {
			return -compareVersionTail(nil, leftTokens[i:])
		}

		cmp := compareVersionToken(leftTokens[i], rightTokens[i])
		if cmp != 0 {
			return cmp
		}
	}

	return 0
}

func compareVersionTail(_ []versionToken, remain []versionToken) int {
	if len(remain) == 0 {
		return 0
	}

	for _, token := range remain {
		if token.isNumber {
			if token.raw != "" && token.raw != "0" {
				return -1
			}
			continue
		}
		return 1
	}

	return 0
}

type versionToken struct {
	raw      string
	isNumber bool
}

func tokenizeVersion(version string) []versionToken {
	version = strings.ToLower(strings.TrimSpace(version))
	version = strings.TrimPrefix(version, "v")
	segmentRegexp := regexp.MustCompile(`[0-9]+|[a-z]+`)
	parts := segmentRegexp.FindAllString(version, -1)
	tokens := make([]versionToken, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		tokens = append(tokens, versionToken{raw: part, isNumber: part[0] >= '0' && part[0] <= '9'})
	}
	if len(tokens) == 0 {
		tokens = append(tokens, versionToken{raw: version, isNumber: false})
	}
	return tokens
}

func compareVersionToken(left versionToken, right versionToken) int {
	if left.isNumber && right.isNumber {
		leftTrim := strings.TrimLeft(left.raw, "0")
		rightTrim := strings.TrimLeft(right.raw, "0")
		if leftTrim == "" {
			leftTrim = "0"
		}
		if rightTrim == "" {
			rightTrim = "0"
		}
		if len(leftTrim) < len(rightTrim) {
			return -1
		}
		if len(leftTrim) > len(rightTrim) {
			return 1
		}
		if leftTrim < rightTrim {
			return -1
		}
		if leftTrim > rightTrim {
			return 1
		}
		return 0
	}

	if left.isNumber && !right.isNumber {
		return 1
	}
	if !left.isNumber && right.isNumber {
		return -1
	}

	if left.raw < right.raw {
		return -1
	}
	if left.raw > right.raw {
		return 1
	}

	return 0
}

func mergeFinding(left SecurityScanFinding, right SecurityScanFinding) SecurityScanFinding {
	if left.Version == "" {
		left.Version = right.Version
	}

	if len(right.CVEs) == 0 {
		return left
	}

	seen := map[string]struct{}{}
	for _, cve := range left.CVEs {
		seen[cve.Code] = struct{}{}
	}
	for _, cve := range right.CVEs {
		if _, ok := seen[cve.Code]; ok {
			continue
		}
		seen[cve.Code] = struct{}{}
		left.CVEs = append(left.CVEs, cve)
	}

	return normalizeUnknownFinding(left)
}

func normalizeUnknownFinding(finding SecurityScanFinding) SecurityScanFinding {
	if strings.TrimSpace(finding.Product) == "" {
		finding.Product = "Unknown"
	}
	if strings.TrimSpace(finding.Vendor) == "" {
		finding.Vendor = "Unknown"
	}
	if strings.TrimSpace(finding.Version) == "" {
		finding.Version = "Unknown"
	}
	if strings.TrimSpace(finding.Severity) == "" {
		finding.Severity = "Unknown"
	}
	return finding
}

func buildCVEMap(cves []CVE) map[string][]CVE {
	cveMap := make(map[string][]CVE)
	for _, cve := range cves {
		name := strings.TrimSpace(cve.Name)
		if name == "" {
			continue
		}

		cveMap[name] = append(cveMap[name], cve)
	}

	return cveMap
}

func buildFingerprintList(fingerprints []Fingerprint) []Fingerprint {
	result := make([]Fingerprint, 0, len(fingerprints))
	for _, fingerprint := range fingerprints {
		if strings.TrimSpace(fingerprint.Name) == "" {
			continue
		}

		result = append(result, fingerprint)
	}

	return result
}

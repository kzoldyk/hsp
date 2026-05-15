package cmd

import (
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

func ResolveVariables(input string, vars map[string]string) (string, []string) {
	matches := varPattern.FindAllStringSubmatch(input, -1)
	result := input
	missing := []string{}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		varName := match[1]
		value, exists := vars[varName]
		if !exists {
			value, exists = vars[strings.ToUpper(strings.ToLower(varName))]
		}
		if !exists {
			missing = append(missing, varName)
			continue
		}
		result = strings.Replace(result, match[0], value, -1)
	}

	return result, missing
}

func ResolveAll(req *RequestBuilder, vars map[string]string) []string {
	allMissing := []string{}

	urlMissing := []string{}
	req.URL, urlMissing = ResolveVariables(req.URL, vars)
	allMissing = append(allMissing, urlMissing...)

	newHeaders := make(map[string]string)
	for key, value := range req.Headers {
		resolvedKey, missing := ResolveVariables(key, vars)
		resolvedValue, missing2 := ResolveVariables(value, vars)
		newHeaders[resolvedKey] = resolvedValue
		allMissing = append(allMissing, missing...)
		allMissing = append(allMissing, missing2...)
	}
	req.Headers = newHeaders

	newParams := make(map[string]string)
	for key, value := range req.QueryParams {
		resolvedKey, missing := ResolveVariables(key, vars)
		resolvedValue, missing2 := ResolveVariables(value, vars)
		newParams[resolvedKey] = resolvedValue
		allMissing = append(allMissing, missing...)
		allMissing = append(allMissing, missing2...)
	}
	req.QueryParams = newParams

	bodyMissing := []string{}
	req.Body, bodyMissing = ResolveVariables(req.Body, vars)
	allMissing = append(allMissing, bodyMissing...)

	seen := make(map[string]bool)
	uniqueMissing := []string{}
	for _, m := range allMissing {
		if m != "" && !seen[m] {
			seen[m] = true
			uniqueMissing = append(uniqueMissing, m)
		}
	}

	return uniqueMissing
}
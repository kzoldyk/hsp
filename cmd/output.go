package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	ColorBoxBorder       = color.New(color.FgCyan)
	ColorSectionTitle    = color.New(color.FgWhite, color.Bold)
	ColorMethodGET       = color.New(color.FgCyan, color.Bold)
	ColorMethodPOST      = color.New(color.FgGreen, color.Bold)
	ColorMethodPUT       = color.New(color.FgYellow, color.Bold)
	ColorMethodPATCH    = color.New(color.FgMagenta, color.Bold)
	ColorMethodDELETE    = color.New(color.FgRed, color.Bold)
	ColorStatus2xx       = color.New(color.FgGreen, color.Bold)
	ColorStatus3xx       = color.New(color.FgYellow, color.Bold)
	ColorStatus4xx       = color.New(color.FgRed, color.Bold)
	ColorStatus5xx       = color.New(color.FgMagenta, color.Bold)
	ColorError           = color.New(color.FgRed, color.Bold)
	ColorSuccess         = color.New(color.FgGreen, color.Bold)
	ColorURL             = color.New(color.FgCyan)
	ColorURLUnresolved   = color.New(color.FgYellow, color.Faint)
	ColorHeaderKey       = color.New(color.FgYellow, color.Bold)
	ColorHeaderValue    = color.New(color.FgWhite, color.Faint)
	ColorParamKey       = color.New(color.FgYellow)
	ColorParamValue     = color.New(color.FgWhite)
	ColorBodyKey        = color.New(color.FgMagenta, color.Bold)
	ColorBodyValue      = color.New(color.FgWhite)
	ColorResponseKey    = color.New(color.FgCyan, color.Bold)
	ColorResponseValue  = color.New(color.FgWhite)
	ColorMetadata       = color.New(color.FgBlack, color.Faint)
	ColorPrompt         = color.New(color.FgWhite, color.Bold)
	ColorVariable       = color.New(color.FgYellow, color.Bold, color.Underline)
)

func GetMethodColor(method string) *color.Color {
	switch strings.ToUpper(method) {
	case "GET":
		return ColorMethodGET
	case "POST":
		return ColorMethodPOST
	case "PUT":
		return ColorMethodPUT
	case "PATCH":
		return ColorMethodPATCH
	case "DELETE":
		return ColorMethodDELETE
	default:
		return ColorMethodGET
	}
}

func GetStatusColor(statusCode int) *color.Color {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return ColorStatus2xx
	case statusCode >= 300 && statusCode < 400:
		return ColorStatus3xx
	case statusCode >= 400 && statusCode < 500:
		return ColorStatus4xx
	case statusCode >= 500 && statusCode < 600:
		return ColorStatus5xx
	default:
		return ColorStatus2xx
	}
}

func DrawBox(title string, width int) string {
	if width < len(title)+4 {
		width = len(title) + 4
	}
	usableWidth := width - 4
	titleLen := len(title)
	if titleLen > usableWidth {
		title = title[:usableWidth]
		titleLen = usableWidth
	}

	halfWidth := (usableWidth - titleLen) / 2
	left := halfWidth
	right := usableWidth - titleLen - halfWidth

	return "+" + strings.Repeat("-", width-2) + "+\n" +
		"|" + strings.Repeat(" ", left) + title + strings.Repeat(" ", right) + "|\n" +
		"+" + strings.Repeat("-", width-2) + "+"
}

func DrawSection(title string, width int) string {
	if width < len(title)+4 {
		width = len(title) + 4
	}
	usableWidth := width - 4
	titleLen := len(title)
	if titleLen > usableWidth {
		title = title[:usableWidth]
		titleLen = usableWidth
	}

	halfWidth := (usableWidth - titleLen) / 2
	left := halfWidth
	right := usableWidth - titleLen - halfWidth

	return "+" + strings.Repeat("-", width-2) + "+\n" +
		"|" + strings.Repeat(" ", left) + title + strings.Repeat(" ", right) + "|\n" +
		"+" + strings.Repeat("-", width-2) + "+"
}

func DrawDoubleBox(title string, width int) string {
	if width < len(title)+4 {
		width = len(title) + 4
	}
	usableWidth := width - 4
	titleLen := len(title)
	if titleLen > usableWidth {
		title = title[:usableWidth]
		titleLen = usableWidth
	}

	halfWidth := (usableWidth - titleLen) / 2
	left := halfWidth
	right := usableWidth - titleLen - halfWidth

	return "+" + strings.Repeat("=", width-2) + "+\n" +
		"|" + strings.Repeat(" ", left) + title + strings.Repeat(" ", right) + "|\n" +
		"+" + strings.Repeat("=", width-2) + "+"
}

func Truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width < 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}

func Pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func RenderRequest(req *RequestBuilder) string {
	width := 80

	var lines []string
	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")

	methodColor := GetMethodColor(req.Method)
	methodStr := methodColor.Sprint(req.Method)
	leftPad := width - 12 - len(req.Method)
	lines = append(lines, "|  REQUEST"+Pad("", leftPad)+"["+methodStr+"]")

	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")

	urlColor := ColorURL
	lines = append(lines, "|  URL       : "+urlColor.Sprint(Truncate(req.URL, width-14)))

	if req.ShowResolved && req.origURL != "" && req.origURL != req.URL {
		lines = append(lines, "|  Resolved  : "+ColorURLUnresolved.Sprint(Truncate(req.origURL, width-14)))
	}

	if len(req.Headers) > 0 {
		first := true
		for key, value := range req.Headers {
			if key == "Authorization" || key == "authorization" {
				value = "***"
			}
			if first {
				lines = append(lines, "|  Headers  : "+ColorHeaderKey.Sprint(key+": ")+ColorHeaderValue.Sprint(Truncate(value, width-26-len(key))))
				first = false
			} else {
				lines = append(lines, "|  "+Pad("", 12)+ColorHeaderKey.Sprint(key+": ")+ColorHeaderValue.Sprint(Truncate(value, width-26-len(key))))
			}
		}
	}

	if len(req.QueryParams) > 0 {
		first := true
		for key, value := range req.QueryParams {
			if first {
				lines = append(lines, "|  Params   : "+ColorParamKey.Sprint(Truncate(key+"=", width-26))+ColorParamValue.Sprint(Truncate(value, width-26-len(key)-1)))
				first = false
			} else {
				lines = append(lines, "|  "+Pad("", 12)+ColorParamKey.Sprint(Truncate(key+"=", width-26))+ColorParamValue.Sprint(Truncate(value, width-26-len(key)-1)))
			}
		}
	}

	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")

	if req.Body != "" {
		lines = append(lines, "|  BODY (Payload)")
		lines = append(lines, "|  "+formatBodyBox(req.Body, width-4))
	} else {
		lines = append(lines, "|  BODY (empty)")
	}

	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")
	return strings.Join(lines, "\n")
}

func formatBodyBox(body string, width int) string {
	if body == "" {
		body = "(empty)"
	}

	rows := strings.Split(body, "\n")
	var lines []string
	for _, line := range rows {
		line = strings.TrimRight(line, " ")
		if len(line) > width-4 {
			line = Truncate(line, width-4)
		}
		lines = append(lines, "| "+Pad(line, width-4)+" |")
	}
	return strings.Join([]string{
		"+" + strings.Repeat("-", width-2) + "+",
		strings.Join(lines, "\n"),
		"+" + strings.Repeat("-", width-2) + "+",
	}, "\n")
}

func RenderRequestPreview(req *RequestBuilder) string {
	width := 60
	methodColor := GetMethodColor(req.Method)
	methodStr := methodColor.Sprint(req.Method)
	urlStr := ColorURL.Sprint(Truncate(req.URL, width-16))

	return "+" + strings.Repeat("-", width-2) + "+\n" +
		"| " + methodStr + " " + urlStr + " |\n" +
		"+" + strings.Repeat("-", width-2) + "+"
}

func (rb *RequestBuilder) RenderRequestPreview() string {
	return RenderRequestPreview(rb)
}

func RenderResponse(statusCode int, statusText string, duration time.Duration, headers http.Header, body []byte) string {
	width := 80

	var lines []string

	timeStr := ColorMetadata.Sprint(duration)
	statusColor := GetStatusColor(statusCode)
	statusStr := statusColor.Sprint(fmt.Sprintf("%d %s", statusCode, statusText))

	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")
	statusLen := len(fmt.Sprintf("%d %s", statusCode, statusText))
	lines = append(lines, "|  Time: "+timeStr+Pad("", width-20-statusLen)+"["+statusStr+"]")
	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")

	lines = append(lines, "|  RESPONSE")

	if len(headers) > 0 {
		first := true
		for key, values := range headers {
			valStr := ""
			for _, v := range values {
				valStr += v
			}
			if first {
				lines = append(lines, "|  Headers  : "+ColorHeaderKey.Sprint(key+": ")+ColorHeaderValue.Sprint(Truncate(valStr, width-26-len(key))))
				first = false
			} else {
				lines = append(lines, "|  "+Pad("", 12)+ColorHeaderKey.Sprint(key+": ")+ColorHeaderValue.Sprint(Truncate(valStr, width-26-len(key))))
			}
		}
	}

	lines = append(lines, "|  "+formatResponseBodyBox(body, width-4))

	lines = append(lines, "+"+strings.Repeat("-", width-2)+"+")
	return strings.Join(lines, "\n")
}

func (rb *RequestBuilder) RenderResponse(statusCode int, statusText string, duration time.Duration, headers http.Header, body []byte) string {
	return RenderResponse(statusCode, statusText, duration, headers, body)
}

func formatResponseBodyBox(body []byte, width int) string {
	if len(body) == 0 {
		body = []byte("(empty)")
	}

	var jsonBody interface{}
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		formatted, err := json.MarshalIndent(jsonBody, "", "  ")
		if err == nil {
			body = formatted
		}
	}

	rows := strings.Split(string(body), "\n")
	var lines []string
	displayCount := 0
	maxLines := 8
	for _, line := range rows {
		line = strings.TrimRight(line, " ")
		if len(line) > width-4 {
			line = Truncate(line, width-4)
		}
		lines = append(lines, "| "+Pad(line, width-4)+" |")
		displayCount++
		if displayCount >= maxLines && len(rows) > maxLines {
			lines = append(lines, "| "+Pad("... and "+fmt.Sprintf("%d", len(rows)-maxLines)+" more lines", width-4)+" |")
			break
		}
	}

	return strings.Join([]string{
		"+" + strings.Repeat("-", width-2) + "+",
		strings.Join(lines, "\n"),
		"+" + strings.Repeat("-", width-2) + "+",
	}, "\n")
}
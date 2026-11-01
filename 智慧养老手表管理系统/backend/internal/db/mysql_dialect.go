package db

import (
	"errors"
	"regexp"
	"strings"
)

var intervals = regexp.MustCompile(`(?i)^INTERVAL\s+'(\d+)\s+(seconds?|minutes?|hours?|days?)'`)
var hardcodedTimeOffset = regexp.MustCompile(`(?i)^now\(\)\s*\+\s*\(\s*\$(\d+)\s*\*\s*interval\s+'1 second'\s*\)`)

var conflictUpdate = regexp.MustCompile(`(?i)ON\s+CONFLICT\s*\([^)]*\)\s+DO\s+UPDATE\s+SET`)
var conflictNothing = regexp.MustCompile(`(?i)ON\s+CONFLICT(?:\s*\([^)]*\))?\s+DO\s+NOTHING`)
var excludedColumn = regexp.MustCompile(`(?i)excluded\.([a-z][a-z0-9_]*)`)
var firstInsertColumn = regexp.MustCompile(`(?is)^\s*INSERT\s+INTO\s+(?:\x60[^\x60]+\x60|"[^"]+"|[a-z_][a-z0-9_]*)(?:\s*\.\s*(?:\x60[^\x60]+\x60|"[^"]+"|[a-z_][a-z0-9_]*))?\s*\(\s*(\x60[^\x60]+\x60|"[^"]+"|[a-z_][a-z0-9_]*)`)
var lockLimit = regexp.MustCompile(`(?i)FOR UPDATE SKIP LOCKED\s+LIMIT\s+(\d+)`)
var dateFormat = regexp.MustCompile(`(?i)^TO_CHAR\(([a-z_.]+),'(YYYY-MM-DD(?: HH24:MI:SS(?:\.MS)?)?)'\)`)
var varcharCast = regexp.MustCompile(`(?i)\bAS varchar\b`)

func rewriteMySQLSyntax(query string) (string, error) {
	var err error
	query, err = rewriteDialectExpressions(query)
	if err != nil {
		return "", err
	}
	query, err = rewriteSQLCodeSegments(query, func(code string) (string, error) {
		code = varcharCast.ReplaceAllString(code, "AS CHAR")
		code = lockLimit.ReplaceAllString(code, "LIMIT $1 FOR UPDATE SKIP LOCKED")
		code = conflictUpdate.ReplaceAllString(code, "ON DUPLICATE KEY UPDATE")
		code = excludedColumn.ReplaceAllString(code, "VALUES($1)")
		if conflictNothing.MatchString(code) {
			match := firstInsertColumn.FindStringSubmatch(query)
			if len(match) != 2 {
				return "", errors.New("ON CONFLICT DO NOTHING requires an explicit INSERT column list for MySQL translation")
			}
			column := strings.Trim(match[1], "`\"")
			if column == "" {
				return "", errors.New("cannot determine MySQL no-op conflict column")
			}
			code = conflictNothing.ReplaceAllString(code, "ON DUPLICATE KEY UPDATE `"+column+"`=`"+column+"`")
		}
		return code, nil
	})
	if err != nil {
		return "", err
	}
	return query, nil
}

func rewriteSQLCodeSegments(query string, rewrite func(string) (string, error)) (string, error) {
	var out strings.Builder
	codeStart := 0
	for i := 0; i < len(query); {
		if strings.HasPrefix(query[i:], "--") || query[i] == '#' {
			code, err := rewrite(query[codeStart:i])
			if err != nil {
				return "", err
			}
			out.WriteString(code)
			end := strings.IndexByte(query[i:], '\n')
			if end < 0 {
				out.WriteString(query[i:])
				return out.String(), nil
			}
			end += i + 1
			out.WriteString(query[i:end])
			i, codeStart = end, end
			continue
		}
		if strings.HasPrefix(query[i:], "/*") {
			code, err := rewrite(query[codeStart:i])
			if err != nil {
				return "", err
			}
			out.WriteString(code)
			end := strings.Index(query[i+2:], "*/")
			if end < 0 {
				return "", errors.New("unterminated SQL block comment")
			}
			end += i + 4
			out.WriteString(query[i:end])
			i, codeStart = end, end
			continue
		}
		if query[i] == '\'' || query[i] == '"' || query[i] == '`' {
			code, err := rewrite(query[codeStart:i])
			if err != nil {
				return "", err
			}
			out.WriteString(code)
			quote := query[i]
			start := i
			i++
			closed := false
			for i < len(query) {
				c := query[i]
				i++
				if c == '\\' && i < len(query) {
					i++
					continue
				}
				if c == quote {
					if i < len(query) && query[i] == quote {
						i++
						continue
					}
					closed = true
					break
				}
			}
			if !closed {
				return "", errors.New("unterminated SQL quoted value")
			}
			out.WriteString(query[start:i])
			codeStart = i
			continue
		}
		i++
	}
	code, err := rewrite(query[codeStart:])
	if err != nil {
		return "", err
	}
	out.WriteString(code)
	return out.String(), nil
}

func rewriteDialectExpressions(query string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(query); {
		if strings.HasPrefix(query[i:], "--") || query[i] == '#' {
			end := strings.IndexByte(query[i:], '\n')
			if end < 0 {
				out.WriteString(query[i:])
				break
			}
			end += i + 1
			out.WriteString(query[i:end])
			i = end
			continue
		}
		if strings.HasPrefix(query[i:], "/*") {
			end := strings.Index(query[i+2:], "*/")
			if end < 0 {
				return "", errors.New("unterminated SQL block comment")
			}
			end += i + 4
			out.WriteString(query[i:end])
			i = end
			continue
		}
		if query[i] == '\'' || query[i] == '"' || query[i] == '`' {
			quote, start := query[i], i
			i++
			closed := false
			for i < len(query) {
				c := query[i]
				i++
				if c == '\\' && i < len(query) {
					i++
					continue
				}
				if c == quote {
					if i < len(query) && query[i] == quote {
						i++
						continue
					}
					closed = true
					break
				}
			}
			if !closed {
				return "", errors.New("unterminated SQL quoted value")
			}
			out.WriteString(query[start:i])
			continue
		}
		if i == 0 || !identifierByte(query[i-1]) {
			if match := hardcodedTimeOffset.FindStringSubmatch(query[i:]); match != nil {
				out.WriteString("TIMESTAMPADD(SECOND,$" + match[1] + ",now())")
				i += len(match[0])
				continue
			}
			if match := dateFormat.FindStringSubmatch(query[i:]); match != nil {
				format := "%Y-%m-%d"
				if strings.Contains(strings.ToUpper(match[2]), "HH24") {
					format += " %H:%i:%s"
				}
				if strings.HasSuffix(strings.ToUpper(match[2]), ".MS") {
					out.WriteString("LEFT(DATE_FORMAT(" + match[1] + ",'" + format + ".%f'),23)")
				} else {
					out.WriteString("DATE_FORMAT(" + match[1] + ",'" + format + "')")
				}
				i += len(match[0])
				continue
			}
			if match := intervals.FindStringSubmatch(query[i:]); match != nil {
				out.WriteString("INTERVAL " + match[1] + " " + strings.TrimSuffix(strings.ToUpper(match[2]), "S"))
				i += len(match[0])
				continue
			}
		}
		out.WriteByte(query[i])
		i++
	}
	return out.String(), nil
}

func containsSQLCodePattern(query string, pattern *regexp.Regexp) (bool, error) {
	found := false
	_, err := rewriteSQLCodeSegments(query, func(code string) (string, error) {
		if pattern.MatchString(code) {
			found = true
		}
		return code, nil
	})
	return found, err
}

// Package stalog provides application logger for Cloud Logging.
package stalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type AdditionalData map[string]interface{}

// Config is the configuration for `RequestLogging` middleware.
type Config struct {
	ProjectId string

	// Output for request log
	RequestLogOut io.Writer

	// Output for context log (application log)
	ContextLogOut io.Writer

	Severity       Severity
	AdditionalData AdditionalData

	// nest level for runtime.Caller (default: 2)
	Skip int
}

// NewConfig creates a config with default settings.
func NewConfig(projectId string) *Config {
	return &Config{
		ProjectId:      projectId,
		Severity:       SeverityInfo,
		RequestLogOut:  os.Stderr,
		ContextLogOut:  os.Stdout,
		AdditionalData: AdditionalData{},
		Skip:           2,
	}
}

// Severity is the level of log. More details:
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
type Severity int

const (
	SeverityDefault Severity = iota * 100
	SeverityDebug
	SeverityInfo
	SeverityNotice
	SeverityWarning
	SeverityError
	SeverityCritical
	SeverityAlert
	SeverityEmergency
)

// String returns text representation for the severity
func (s Severity) String() string {
	switch s {
	case SeverityDefault:
		return "DEFAULT"
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityNotice:
		return "NOTICE"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityAlert:
		return "ALERT"
	case SeverityEmergency:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

type SourceLocation struct {
	File     string `json:"file"`
	Line     string `json:"line"`
	Function string `json:"function"`
}

type contextLog struct {
	Time           string         `json:"time"`
	Trace          string         `json:"logging.googleapis.com/trace"`
	SourceLocation SourceLocation `json:"logging.googleapis.com/sourceLocation"`
	Severity       string         `json:"severity"`
	Message        string         `json:"message"`
	AdditionalData AdditionalData `json:"data,omitempty"`
}

// ContextLogger is the logger which is combined with the request
type ContextLogger struct {
	out            io.Writer
	Trace          string
	Severity       Severity
	AdditionalData AdditionalData
	loggedSeverity []Severity
	Skip           int
}

// RequestContextLogger gets request-context logger for the request.
// You must use `RequestLogging` middleware in advance for this function to work.
func RequestContextLogger(r *http.Request) *ContextLogger {
	v, _ := r.Context().Value(ContextLoggerKey).(*ContextLogger)
	return v
}

// Default logs a message at DEFAULT severity
func (l *ContextLogger) Default(args ...interface{}) {
	_ = l.write(SeverityDefault, fmt.Sprint(args...))
}

// Defaultf logs a message at DEFAULT severity
func (l *ContextLogger) Defaultf(format string, args ...interface{}) {
	_ = l.write(SeverityDefault, fmt.Sprintf(format, args...))
}

// Defaultln logs a message at DEFAULT severity
func (l *ContextLogger) Defaultln(args ...interface{}) {
	_ = l.write(SeverityDefault, fmt.Sprintln(args...))
}

// Debug logs a message at DEBUG severity
func (l *ContextLogger) Debug(args ...interface{}) {
	_ = l.write(SeverityDebug, fmt.Sprint(args...))
}

// Debugf logs a message at DEBUG severity
func (l *ContextLogger) Debugf(format string, args ...interface{}) {
	_ = l.write(SeverityDebug, fmt.Sprintf(format, args...))
}

// Debugln logs a message at DEBUG severity
func (l *ContextLogger) Debugln(args ...interface{}) {
	_ = l.write(SeverityDebug, fmt.Sprintln(args...))
}

// Info logs a message at INFO severity
func (l *ContextLogger) Info(args ...interface{}) {
	_ = l.write(SeverityInfo, fmt.Sprint(args...))
}

// Infof logs a message at INFO severity
func (l *ContextLogger) Infof(format string, args ...interface{}) {
	_ = l.write(SeverityInfo, fmt.Sprintf(format, args...))
}

// Infoln logs a message at INFO severity
func (l *ContextLogger) Infoln(args ...interface{}) {
	_ = l.write(SeverityInfo, fmt.Sprintln(args...))
}

// Notice logs a message at NOTICE severity
func (l *ContextLogger) Notice(args ...interface{}) {
	_ = l.write(SeverityNotice, fmt.Sprint(args...))
}

// Noticef logs a message at NOTICE severity
func (l *ContextLogger) Noticef(format string, args ...interface{}) {
	_ = l.write(SeverityNotice, fmt.Sprintf(format, args...))
}

// Noticeln logs a message at NOTICE severity
func (l *ContextLogger) Noticeln(args ...interface{}) {
	_ = l.write(SeverityNotice, fmt.Sprintln(args...))
}

// Warning logs a message at WARNING severity
func (l *ContextLogger) Warning(args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprint(args...))
}

// Warningf logs a message at WARNING severity
func (l *ContextLogger) Warningf(format string, args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprintf(format, args...))
}

// Warningln logs a message at WARNING severity
func (l *ContextLogger) Warningln(args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprintln(args...))
}

// Warn logs a message at WARNING severity
func (l *ContextLogger) Warn(args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprint(args...))
}

// Warnf logs a message at WARNING severity
func (l *ContextLogger) Warnf(format string, args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprintf(format, args...))
}

// Warnln logs a message at WARNING severity
func (l *ContextLogger) Warnln(args ...interface{}) {
	_ = l.write(SeverityWarning, fmt.Sprintln(args...))
}

// Error logs a message at ERROR severity
func (l *ContextLogger) Error(args ...interface{}) {
	_ = l.write(SeverityError, fmt.Sprint(args...))
}

// Errorf logs a message at ERROR severity
func (l *ContextLogger) Errorf(format string, args ...interface{}) {
	_ = l.write(SeverityError, fmt.Sprintf(format, args...))
}

// Errorln logs a message at ERROR severity
func (l *ContextLogger) Errorln(args ...interface{}) {
	_ = l.write(SeverityError, fmt.Sprintln(args...))
}

// Critical logs a message at CRITICAL severity
func (l *ContextLogger) Critical(args ...interface{}) {
	_ = l.write(SeverityCritical, fmt.Sprint(args...))
}

// Criticalf logs a message at CRITICAL severity
func (l *ContextLogger) Criticalf(format string, args ...interface{}) {
	_ = l.write(SeverityCritical, fmt.Sprintf(format, args...))
}

// Criticalln logs a message at CRITICAL severity
func (l *ContextLogger) Criticalln(args ...interface{}) {
	_ = l.write(SeverityCritical, fmt.Sprintln(args...))
}

// Alert logs a message at ALERT severity
func (l *ContextLogger) Alert(args ...interface{}) {
	_ = l.write(SeverityAlert, fmt.Sprint(args...))
}

// Alertf logs a message at ALERT severity
func (l *ContextLogger) Alertf(format string, args ...interface{}) {
	_ = l.write(SeverityAlert, fmt.Sprintf(format, args...))
}

// Alertln logs a message at ALERT severity
func (l *ContextLogger) Alertln(args ...interface{}) {
	_ = l.write(SeverityAlert, fmt.Sprintln(args...))
}

// Emergency logs a message at EMERGENCY severity
func (l *ContextLogger) Emergency(args ...interface{}) {
	_ = l.write(SeverityEmergency, fmt.Sprint(args...))
}

// Emergencyf logs a message at EMERGENCY severity
func (l *ContextLogger) Emergencyf(format string, args ...interface{}) {
	_ = l.write(SeverityEmergency, fmt.Sprintf(format, args...))
}

// Emergencyln logs a message at EMERGENCY severity
func (l *ContextLogger) Emergencyln(args ...interface{}) {
	_ = l.write(SeverityEmergency, fmt.Sprintln(args...))
}

func (l *ContextLogger) write(severity Severity, msg string) error {
	if severity < l.Severity {
		return nil
	}

	l.loggedSeverity = append(l.loggedSeverity, severity)

	// get source location
	var location SourceLocation
	if pc, file, line, ok := runtime.Caller(l.Skip); ok {
		if function := runtime.FuncForPC(pc); function != nil {
			location.Function = function.Name()
		}
		location.Line = fmt.Sprintf("%d", line)
		parts := strings.Split(file, "/")
		location.File = parts[len(parts)-1] // use short file name
	}

	log := &contextLog{
		Time:           time.Now().Format(time.RFC3339Nano),
		Trace:          l.Trace,
		SourceLocation: location,
		Severity:       severity.String(),
		Message:        msg,
		AdditionalData: l.AdditionalData,
	}

	jsonByte, err := json.Marshal(log)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	// append \n
	jsonByte = append(jsonByte, 0xa)

	_, err = l.out.Write(jsonByte)
	return err
}

func (l *ContextLogger) maxSeverity() Severity {
	max := SeverityDefault
	for _, s := range l.loggedSeverity {
		if s > max {
			max = s
		}
	}

	return max
}

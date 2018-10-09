package stackdriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"go.opencensus.io/exporter/stackdriver/propagation"
)

func Handler(config *Config, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()

		traceId := getTraceId(r)
		trace := fmt.Sprintf("projects/%s/traces/%s", config.projectId, traceId)

		contextLogger := &ContextLogger{
			logger:         config.contextLogger,
			Trace:          trace,
			Severity:       config.severity,
			loggedSeverity: make([]Severity, 0, 10),
		}
		ctx := context.WithValue(r.Context(), contextLoggerKey, contextLogger)
		r = r.WithContext(ctx)

		wrw := &WrappedResponseWriter{ResponseWriter: w}
		defer func() {
			// logging
			elapsed := time.Since(before)
			maxSeverity := contextLogger.maxSeverity()
			err := writeRequestLog(r, config, wrw.status, wrw.responseSize, elapsed, trace, maxSeverity)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}()
		next.ServeHTTP(wrw, r)
	}
	return http.HandlerFunc(fn)
}

func getTraceId(r *http.Request) string {
	httpFormat := &propagation.HTTPFormat{}
	if sc, ok := httpFormat.SpanContextFromRequest(r); ok {
		return sc.TraceID.String()
	} else {
		// TODO
		uniqueBytes := sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano)))
		return hex.EncodeToString(uniqueBytes[:16])
	}
}

type WrappedResponseWriter struct {
	http.ResponseWriter
	status       int
	responseSize int
}

func (w *WrappedResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *WrappedResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.responseSize += n
	return n, err
}

func writeRequestLog(r *http.Request, config *Config, status int, responseSize int, elapsed time.Duration, trace string, severity Severity) error {
	requestLog := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339Nano),
		"logging.googleapis.com/trace": trace,
		"severity":                     severity.String(),
		"httpRequest": map[string]interface{}{
			"requestMethod":                  r.Method,
			"requestUrl":                     r.URL.RequestURI(),
			"requestSize":                    fmt.Sprintf("%d", r.ContentLength),
			"status":                         status,
			"responseSize":                   fmt.Sprintf("%d", responseSize),
			"userAgent":                      r.UserAgent(),
			"remoteIp":                       r.RemoteAddr,
			"serverIp":                       getServerIp(),
			"referer":                        r.Referer(),
			"latency":                        fmt.Sprintf("%fs", elapsed.Seconds()),
			"cacheLookUp":                    false,
			"cacheHit":                       false,
			"cacheValidatedWithOriginServer": false,
			"protocol":                       r.Proto,
		},
		"logType": "request_log",
	}
	for k, v := range config.additional {
		requestLog[k] = v
	}
	requestLogJson, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}

	config.requestLogger.Println(string(requestLogJson))

	return nil
}

func getServerIp() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}

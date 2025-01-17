// Copyright (c) 2023 Snowflake Computing Inc. All rights reserved.

package gosnowflake

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	rlog "github.com/sirupsen/logrus"
)

func createTestLogger() *defaultLogger {
	var rLogger = rlog.New()
	var ret = defaultLogger{inner: rLogger}
	return &ret
}

func TestIsLevelEnabled(t *testing.T) {
	logger := createTestLogger()
	err := logger.SetLogLevel("trace")
	if err != nil {
		t.Fatalf("log level could not be set %v", err)
	}
	if logger.GetLogLevel() != "trace" {
		t.Fatalf("log level should be trace but is %v", logger.GetLogLevel())
	}
}

func TestLogFunction(t *testing.T) {
	logger := createTestLogger()
	buf := &bytes.Buffer{}
	var formatter = rlog.TextFormatter{CallerPrettyfier: SFCallerPrettyfier}
	logger.SetFormatter(&formatter)
	logger.SetReportCaller(true)
	logger.SetOutput(buf)
	err := logger.SetLogLevel("trace")
	if err != nil {
		t.Fatalf("log level could not be set %v", err)
	}

	logger.Trace("hello world")
	logger.Tracef("log %v", "format")
	logger.Traceln("log line")

	var strbuf = buf.String()
	if !strings.Contains(strbuf, "hello world") &&
		!strings.Contains(strbuf, "log format") &&
		!strings.Contains(strbuf, "log line") {
		t.Fatalf("unexpected output in log %v", strbuf)
	}
}

func TestSetLogLevelError(t *testing.T) {
	logger := CreateDefaultLogger()
	err := logger.SetLogLevel("unknown")
	if err == nil {
		t.Fatal("should have thrown an error")
	}
}

func TestDefaultLogLevel(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	SetLogger(&logger)

	// default logger level is info
	logger.Info("info")
	logger.Infof("info%v", "f")
	logger.Infoln("infoln")

	// debug and trace won't write to log since they are higher than info level
	logger.Debug("debug")
	logger.Debugf("debug%v", "f")
	logger.Debugln("debugln")

	logger.Trace("trace")
	logger.Tracef("trace%v", "f")
	logger.Traceln("traceln")

	// print, warning and error should write to log since they are lower than info
	logger.Print("print")
	logger.Printf("print%v", "f")
	logger.Println("println")

	logger.Warn("warn")
	logger.Warnf("warn%v", "f")
	logger.Warnln("warnln")

	logger.Warning("warning")
	logger.Warningf("warning%v", "f")
	logger.Warningln("warningln")

	logger.Error("error")
	logger.Errorf("error%v", "f")
	logger.Errorln("errorln")

	// verify output
	var strbuf = buf.String()

	if strings.Contains(strbuf, "debug") &&
		strings.Contains(strbuf, "trace") &&
		!strings.Contains(strbuf, "info") &&
		!strings.Contains(strbuf, "print") &&
		!strings.Contains(strbuf, "warn") &&
		!strings.Contains(strbuf, "warning") &&
		!strings.Contains(strbuf, "error") {
		t.Fatalf("unexpected output in log: %v", strbuf)
	}
}

func TestOffLogLevel(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	err := logger.SetLogLevel("OFF")
	assertNilF(t, err)
	SetLogger(&logger)

	logger.Info("info")
	logger.Infof("info%v", "f")
	logger.Infoln("infoln")
	logger.Debug("debug")
	logger.Debugf("debug%v", "f")
	logger.Debugln("debugln")
	logger.Trace("trace")
	logger.Tracef("trace%v", "f")
	logger.Traceln("traceln")
	logger.Print("print")
	logger.Printf("print%v", "f")
	logger.Println("println")
	logger.Warn("warn")
	logger.Warnf("warn%v", "f")
	logger.Warnln("warnln")
	logger.Warning("warning")
	logger.Warningf("warning%v", "f")
	logger.Warningln("warningln")
	logger.Error("error")
	logger.Errorf("error%v", "f")
	logger.Errorln("errorln")

	assertEqualE(t, buf.Len(), 0, "log messages count")
	assertEqualE(t, logger.GetLogLevel(), "OFF", "log level")
}

func TestLogSetLevel(t *testing.T) {
	logger := GetLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	_ = logger.SetLogLevel("trace")

	logger.Trace("should print at trace level")
	logger.Debug("should print at debug level")

	var strbuf = buf.String()

	if !strings.Contains(strbuf, "trace level") &&
		!strings.Contains(strbuf, "debug level") {
		t.Fatalf("unexpected output in log: %v", strbuf)
	}
}

func TestLogWithError(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	err := errors.New("error")
	logger.WithError(err).Info("hello world")

	var strbuf = buf.String()
	if !strings.Contains(strbuf, "error=error") {
		t.Fatalf("unexpected output in log: %v", strbuf)
	}
}

func TestLogWithTime(t *testing.T) {
	logger := createTestLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ti := time.Now()
	logger.WithTime(ti).Info("hello")
	time.Sleep(3 * time.Second)

	var strbuf = buf.String()
	if !strings.Contains(strbuf, ti.Format(time.RFC3339)) {
		t.Fatalf("unexpected string in output: %v", strbuf)
	}
}

func TestLogWithField(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	logger.WithField("field", "test").Info("hello")
	var strbuf = buf.String()
	if !strings.Contains(strbuf, "field=test") {
		t.Fatalf("unexpected string in output: %v", strbuf)
	}
}

type testRequestIDCtxKey struct{}

func TestLogKeysDefault(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()

	// set the sessionID on the context to see if we have it in the logs
	sessionIDContextValue := "sessionID"
	ctx = context.WithValue(ctx, SFSessionIDKey, sessionIDContextValue)

	userContextValue := "madison"
	ctx = context.WithValue(ctx, SFSessionUserKey, userContextValue)

	// base case (not using RegisterContextVariableToLog to add additional types )
	logger.WithContext(ctx).Info("test")
	var strbuf = buf.String()
	if !strings.Contains(strbuf, fmt.Sprintf("%s=%s", SFSessionIDKey, sessionIDContextValue)) {
		t.Fatalf("expected that sfSessionIdKey would be in logs if logger.WithContext was used, but got: %v", strbuf)
	}
	if !strings.Contains(strbuf, fmt.Sprintf("%s=%s", SFSessionUserKey, userContextValue)) {
		t.Fatalf("expected that SFSessionUserKey would be in logs if logger.WithContext was used, but got: %v", strbuf)
	}
}

func TestLogKeysWithRegisterContextVariableToLog(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()

	// set the sessionID on the context to see if we have it in the logs
	sessionIDContextValue := "sessionID"
	ctx = context.WithValue(ctx, SFSessionIDKey, sessionIDContextValue)

	userContextValue := "testUser"
	ctx = context.WithValue(ctx, SFSessionUserKey, userContextValue)

	// test that RegisterContextVariableToLog works with non string keys
	logKey := "REQUEST_ID"
	contextIntVal := 123
	ctx = context.WithValue(ctx, testRequestIDCtxKey{}, contextIntVal)

	getRequestKeyFunc := func(ctx context.Context) string {
		if requestContext, ok := ctx.Value(testRequestIDCtxKey{}).(int); ok {
			return fmt.Sprint(requestContext)
		}
		return ""
	}

	RegisterLogContextHook(logKey, getRequestKeyFunc)

	// base case (not using RegisterContextVariableToLog to add additional types )
	logger.WithContext(ctx).Info("test")
	var strbuf = buf.String()

	if !strings.Contains(strbuf, fmt.Sprintf("%s=%s", SFSessionIDKey, sessionIDContextValue)) {
		t.Fatalf("expected that sfSessionIdKey would be in logs if logger.WithContext and RegisterContextVariableToLog was used, but got: %v", strbuf)
	}
	if !strings.Contains(strbuf, fmt.Sprintf("%s=%s", SFSessionUserKey, userContextValue)) {
		t.Fatalf("expected that SFSessionUserKey would be in logs if logger.WithContext and RegisterContextVariableToLog was used, but got: %v", strbuf)
	}
	if !strings.Contains(strbuf, fmt.Sprintf("%s=%s", logKey, fmt.Sprint(contextIntVal))) {
		t.Fatalf("expected that REQUEST_ID would be in logs if logger.WithContext and RegisterContextVariableToLog was used, but got: %v", strbuf)
	}
}

func TestLogMaskSecrets(t *testing.T) {
	logger := CreateDefaultLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()
	query := "create user testuser password='testpassword'"
	logger.WithContext(ctx).Infof("Query: %#v", query)

	// verify output
	expected := "create user testuser password='****"
	var strbuf = buf.String()
	if !strings.Contains(strbuf, expected) {
		t.Fatalf("expected that password would be masked. WithContext was used, but got: %v", strbuf)
	}
}

func TestLogEntryConversion(t *testing.T) {
	log := createTestLogger()
	e := log.WithField("k", 1)
	l, ok := e.(ConvertibleEntry)
	if !ok {
		t.Fatalf("expected a ConvertibleEntry")
	}
	v, ok := l.ToEntry().Data["k"]
	if !ok {
		t.Fatalf("should be able to access the underlying rlog.Entry")
	}
	if v.(int) != 1 {
		t.Fatalf("expected value to be 1")
	}
}

package wwlog

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type (
	levelTypeFunc    func(int, string, ...interface{})
	msgTypeFunc      func(string, ...interface{})
	errTypeFunc      func(error, string, ...interface{})
	levelErrTypeFunc func(int, error, string, ...interface{})
)

func Test_Log(t *testing.T) {
	t.Logf("Running Test_Log test")
	SetLogLevel(DEBUG)

	tests := []struct {
		name             string
		msgTypeFunc      msgTypeFunc
		levelTypeFunc    levelTypeFunc
		errTypeFunc      errTypeFunc
		levelErrTypeFunc levelErrTypeFunc
		level            int
		err              error
		message          string
		args             interface{}
		expect           string
		exactMatch       bool
	}{
		{
			name:          "Log DEBUG",
			levelTypeFunc: Log,
			level:         DEBUG,
			message:       "Log",
			expect:        "DEBUG  : Log\n",
			exactMatch:    true,
		},
		{
			name:          "Log ERROR",
			levelTypeFunc: Log,
			level:         ERROR,
			message:       "Log",
			expect:        "ERROR  : Log\n",
			exactMatch:    true,
		},
		{
			name:             "LogExc",
			levelErrTypeFunc: LogExc,
			level:            INFO,
			err:              errors.New("error"),
			message:          "Log",
			expect:           "INFO   : Log\nerror\n",
		},
		{
			name:        "Debug",
			msgTypeFunc: Debug,
			message:     "Debug",
			expect:      "DEBUG  : Debug\n",
			exactMatch:  true,
		},
		{
			name:        "DebugExc",
			errTypeFunc: DebugExc,
			message:     "Debug",
			err:         errors.New("error"),
			expect:      "DEBUG  : Debug\nerror\n",
		},
		{
			name:        "SecDebug",
			msgTypeFunc: SecDebug,
			message:     "Debug",
			expect:      "SECDEBUG   : Debug\n",
			exactMatch:  true,
		},
		{
			name:        "Verbose",
			msgTypeFunc: Verbose,
			message:     "Verbose",
			expect:      "VERBOSE: Verbose\n",
			exactMatch:  true,
		},
		{
			name:        "VerboseExc",
			errTypeFunc: VerboseExc,
			message:     "Verbose",
			err:         errors.New("error"),
			expect:      "VERBOSE: Verbose\nerror\n",
		},
		{
			name:        "SecVerbose",
			msgTypeFunc: SecVerbose,
			message:     "Verbose",
			expect:      "SECVERBOSE : Verbose\n",
			exactMatch:  true,
		},
		{
			name:        "Info",
			msgTypeFunc: Info,
			message:     "Info",
			expect:      "INFO   : Info\n",
			exactMatch:  true,
		},
		{
			name:        "InfoExc",
			errTypeFunc: InfoExc,
			message:     "Info",
			err:         errors.New("error"),
			expect:      "INFO   : Info\nerror\n",
		},
		{
			name:        "SecInfo",
			msgTypeFunc: SecInfo,
			message:     "Info",
			expect:      "SECINFO: Info\n",
			exactMatch:  true,
		},
		{
			name:        "Serv",
			msgTypeFunc: Serv,
			message:     "Serv",
			expect:      "SERV   : Serv\n",
			exactMatch:  true,
		},
		{
			name:        "Recv",
			msgTypeFunc: Recv,
			message:     "Recv",
			expect:      "RECV   : Recv\n",
			exactMatch:  true,
		},
		{
			name:        "Send",
			msgTypeFunc: Send,
			message:     "Send",
			expect:      "SEND   : Send\n",
			exactMatch:  true,
		},
		{
			name:        "Warn",
			msgTypeFunc: Warn,
			message:     "Warn",
			expect:      "WARN   : Warn\n",
			exactMatch:  true,
		},
		{
			name:        "WarnExc",
			errTypeFunc: WarnExc,
			message:     "Warn",
			err:         errors.New("error"),
			expect:      "WARN   : Warn\nerror\n",
		},
		{
			name:        "SecWarn",
			msgTypeFunc: SecWarn,
			message:     "Warn",
			expect:      "SECWARN: Warn\n",
			exactMatch:  true,
		},
		{
			name:        "Error",
			msgTypeFunc: Error,
			message:     "Error",
			expect:      "ERROR  : Error\n",
			exactMatch:  true,
		},
		{
			name:        "ErrorExc",
			errTypeFunc: ErrorExc,
			message:     "Error",
			err:         errors.New("error"),
			expect:      "ERROR  : Error\nerror\n",
		},
		{
			name:        "SecError",
			msgTypeFunc: SecError,
			message:     "Error",
			expect:      "SECERROR   : Error\n",
			exactMatch:  true,
		},
		{
			name:        "Denied",
			msgTypeFunc: Denied,
			message:     "Denied",
			expect:      "DENIED : Denied\n",
			exactMatch:  true,
		},
		{
			name:        "Critical",
			msgTypeFunc: Critical,
			message:     "Critical",
			expect:      "CRITICAL   : Critical\n",
			exactMatch:  true,
		},
		{
			name:        "CriticalExc",
			errTypeFunc: CriticalExc,
			message:     "Critical",
			err:         errors.New("error"),
			expect:      "CRITICAL   : Critical\nerror\n",
		},
		{
			name:        "SecCritical",
			msgTypeFunc: SecCritical,
			message:     "Critical",
			expect:      "SECCRITICAL: Critical\n",
			exactMatch:  true,
		},
	}

	for _, tt := range tests {
		oldErr := os.Stderr

		r, w, err := os.Pipe()
		if err != nil {
			t.Errorf("Could not create stderr pipe, err:%v", err)
			t.FailNow()
		}
		os.Stderr = w
		// make sure os.Stderr is always reset
		defer func() {
			os.Stderr = oldErr
		}()

		SetLogWriter(os.Stderr)

		if tt.msgTypeFunc != nil {
			if tt.args != nil {
				tt.msgTypeFunc(tt.message, tt.args)
			} else {
				tt.msgTypeFunc(tt.message)
			}
		} else if tt.levelTypeFunc != nil {
			if tt.args != nil {
				tt.levelTypeFunc(tt.level, tt.message, tt.args)
			} else {
				tt.levelTypeFunc(tt.level, tt.message)
			}
		} else if tt.errTypeFunc != nil {
			if tt.args != nil {
				tt.errTypeFunc(tt.err, tt.message, tt.args)
			} else {
				tt.errTypeFunc(tt.err, tt.message)
			}
		} else if tt.levelErrTypeFunc != nil {
			if tt.args != nil {
				tt.levelErrTypeFunc(tt.level, tt.err, tt.message, tt.args)
			} else {
				tt.levelErrTypeFunc(tt.level, tt.err, tt.message)
			}
		} else {
			os.Stderr = oldErr
			t.Errorf("One of `msgTypeFunc`, `levelTypeFunc` and `errTypeFunc` should be set")
			t.FailNow()
		}

		outCh := make(chan string, 1)
		go func() {
			out, _ := io.ReadAll(r)
			outCh <- string(out)
		}()

		w.Close()
		os.Stderr = oldErr
		out := <-outCh
		if tt.exactMatch {
			if out != tt.expect {
				t.Errorf("Test: %s failed with unexpected output out: `%s`, expect: `%s`", tt.name, out, tt.expect)
				t.FailNow()
			}
		} else {
			for _, line := range strings.Split(tt.expect, "\n") {
				if !strings.Contains(out, line) {
					t.Errorf("Test: %s should contain output expect: `%s`, out:`%s`", tt.name, line, out)
					t.FailNow()
				}
			}
		}
	}
}

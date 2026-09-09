package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/shirou/gopsutil/v3/process"
)

// withConfig swaps in a check configuration for the duration of fn and restores
// the previous one afterwards. plugin is a package level variable, so tests that
// touch it must not run in parallel.
func withConfig(t *testing.T, cfg Config, fn func()) {
	t.Helper()
	saved := plugin
	plugin = cfg
	plugin.PluginConfig = saved.PluginConfig
	defer func() { plugin = saved }()
	fn()
}

// captureOutput collects everything fn writes to stdout.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = saved }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	return string(out)
}

// ownProcessName is the process name gopsutil reports for the test binary. It is
// used to guarantee that executeCheck finds a matching process.
func ownProcessName(t *testing.T) string {
	t.Helper()
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Skipf("cannot inspect own process: %v", err)
	}
	name, err := p.Name()
	if err != nil || name == "" {
		t.Skipf("cannot determine own process name: %v", err)
	}
	return name
}

func TestCheckArgs(t *testing.T) {
	// A configuration that passes every validation, used as the base for the
	// cases below so each one only varies the field under test.
	valid := Config{
		Process:    "nginx",
		CPUWarn:    50,
		CPUCrit:    75,
		MemoryWarn: 50,
		MemoryCrit: 70,
	}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		want    int
		wantErr string
	}{
		{
			name:   "valid configuration",
			mutate: func(*Config) {},
			want:   sensu.CheckStateOK,
		},
		{
			name:    "process is required",
			mutate:  func(c *Config) { c.Process = "" },
			want:    sensu.CheckStateWarning,
			wantErr: "process is required",
		},
		{
			name:    "cpu-crit of 100 is rejected",
			mutate:  func(c *Config) { c.CPUCrit = 100 },
			want:    sensu.CheckStateWarning,
			wantErr: "that's just stupid",
		},
		{
			name:    "cpu-warn of 100 is rejected",
			mutate:  func(c *Config) { c.CPUWarn = 100 },
			want:    sensu.CheckStateWarning,
			wantErr: "that's just stupid",
		},
		{
			name:    "memory-crit of 100 is rejected",
			mutate:  func(c *Config) { c.MemoryCrit = 100 },
			want:    sensu.CheckStateWarning,
			wantErr: "that's just stupid",
		},
		{
			name:    "memory-warn of 100 is rejected",
			mutate:  func(c *Config) { c.MemoryWarn = 100 },
			want:    sensu.CheckStateWarning,
			wantErr: "that's just stupid",
		},
		{
			name:   "thresholds above 100 are allowed",
			mutate: func(c *Config) { c.CPUCrit = 400 },
			want:   sensu.CheckStateOK,
		},
		{
			name: "time thresholds are optional",
			mutate: func(c *Config) {
				c.TimeWarn = 0
				c.TimeCrit = 0
			},
			want: sensu.CheckStateOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)

			withConfig(t, cfg, func() {
				status, err := checkArgs(nil)
				if status != tt.want {
					t.Errorf("status = %d, want %d", status, tt.want)
				}
				switch {
				case tt.wantErr == "" && err != nil:
					t.Errorf("unexpected error: %v", err)
				case tt.wantErr != "" && err == nil:
					t.Errorf("expected error %q, got nil", tt.wantErr)
				case tt.wantErr != "" && err.Error() != tt.wantErr:
					t.Errorf("error = %q, want %q", err, tt.wantErr)
				}
			})
		})
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		unit float64
		want float64
	}{
		{name: "rounds down to one decimal", x: 12.34, unit: 0.1, want: 12.3},
		{name: "rounds up to one decimal", x: 12.36, unit: 0.1, want: 12.4},
		{name: "rounds half away from zero", x: 12.5, unit: 1, want: 13},
		{name: "rounds to whole units", x: 12.6, unit: 1, want: 13},
		{name: "keeps exact values", x: 12.3, unit: 0.1, want: 12.3},
		{name: "handles zero", x: 0, unit: 0.1, want: 0},
		{name: "handles negatives", x: -12.36, unit: 0.1, want: -12.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Round(tt.x, tt.unit)
			// Compare with a tolerance rather than with ==: neither 12.3 nor
			// 0.1 is exactly representable as a float64.
			if diff := got - tt.want; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("Round(%v, %v) = %v, want %v", tt.x, tt.unit, got, tt.want)
			}
		})
	}
}

func TestExecuteCheckNoMatchingProcess(t *testing.T) {
	// Thresholds of 0 would fire for any process, so an OK result proves the
	// name never matched.
	cfg := Config{
		Process:    "sensu-process-ressources-no-such-process",
		CPUCrit:    0,
		CPUWarn:    0,
		MemoryCrit: 0,
		MemoryWarn: 0,
	}

	withConfig(t, cfg, func() {
		out := captureOutput(t, func() {
			status, err := executeCheck(nil)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if status != sensu.CheckStateOK {
				t.Errorf("status = %d, want %d (OK)", status, sensu.CheckStateOK)
			}
		})
		if out != "" {
			t.Errorf("expected no output, got %q", out)
		}
	})
}

func TestExecuteCheckThresholds(t *testing.T) {
	name := ownProcessName(t)

	// unreachable keeps a resource from firing: memory and CPU percentages are
	// compared with >=, so a threshold above any real value never matches.
	const unreachable = 1e6

	tests := []struct {
		name       string
		cfg        Config
		wantStatus int
		wantOutput string
	}{
		{
			name: "memory critical",
			cfg: Config{
				MemoryCrit: 0,
				MemoryWarn: 0,
				CPUCrit:    unreachable,
				CPUWarn:    unreachable,
			},
			wantStatus: sensu.CheckStateCritical,
			wantOutput: "% memory",
		},
		{
			name: "memory warning",
			cfg: Config{
				MemoryCrit: unreachable,
				MemoryWarn: 0,
				CPUCrit:    unreachable,
				CPUWarn:    unreachable,
			},
			wantStatus: sensu.CheckStateWarning,
			wantOutput: "% memory",
		},
		{
			name: "cpu critical",
			cfg: Config{
				MemoryCrit: unreachable,
				MemoryWarn: unreachable,
				CPUCrit:    0,
				CPUWarn:    0,
			},
			wantStatus: sensu.CheckStateCritical,
			wantOutput: "% CPU",
		},
		{
			name: "cpu warning",
			cfg: Config{
				MemoryCrit: unreachable,
				MemoryWarn: unreachable,
				CPUCrit:    unreachable,
				CPUWarn:    0,
			},
			wantStatus: sensu.CheckStateWarning,
			wantOutput: "% CPU",
		},
		{
			name: "memory is reported before cpu",
			cfg: Config{
				MemoryCrit: 0,
				MemoryWarn: 0,
				CPUCrit:    0,
				CPUWarn:    0,
			},
			wantStatus: sensu.CheckStateCritical,
			wantOutput: "% memory",
		},
		{
			name: "zero time thresholds are disabled",
			cfg: Config{
				MemoryCrit: unreachable,
				MemoryWarn: unreachable,
				CPUCrit:    unreachable,
				CPUWarn:    unreachable,
				TimeCrit:   0,
				TimeWarn:   0,
			},
			wantStatus: sensu.CheckStateOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			cfg.Process = name

			withConfig(t, cfg, func() {
				var status int
				out := captureOutput(t, func() {
					var err error
					status, err = executeCheck(nil)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				})

				if status != tt.wantStatus {
					t.Errorf("status = %d, want %d (output: %q)", status, tt.wantStatus, out)
				}
				if tt.wantOutput == "" {
					if out != "" {
						t.Errorf("expected no output, got %q", out)
					}
					return
				}
				if !strings.Contains(out, tt.wantOutput) {
					t.Errorf("output %q does not contain %q", out, tt.wantOutput)
				}
				if !strings.Contains(out, name) {
					t.Errorf("output %q does not name the process %q", out, name)
				}
			})
		})
	}
}

func TestExecuteCheckRuntimeThresholds(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the test process to age past the time threshold")
	}
	name := ownProcessName(t)

	// The runtime is derived from the process create time rounded to whole
	// seconds, so wait long enough that a threshold of 1 second is exceeded
	// regardless of which way the rounding went.
	time.Sleep(2100 * time.Millisecond)

	const unreachable = 1e6

	tests := []struct {
		name       string
		timeWarn   int64
		timeCrit   int64
		wantStatus int
	}{
		{name: "time critical", timeCrit: 1, wantStatus: sensu.CheckStateCritical},
		{name: "time warning", timeWarn: 1, wantStatus: sensu.CheckStateWarning},
		{name: "critical takes precedence", timeWarn: 1, timeCrit: 1, wantStatus: sensu.CheckStateCritical},
		{name: "threshold not reached", timeCrit: 86400, timeWarn: 86400, wantStatus: sensu.CheckStateOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Process:    name,
				MemoryCrit: unreachable,
				MemoryWarn: unreachable,
				CPUCrit:    unreachable,
				CPUWarn:    unreachable,
				TimeWarn:   tt.timeWarn,
				TimeCrit:   tt.timeCrit,
			}

			withConfig(t, cfg, func() {
				var status int
				out := captureOutput(t, func() {
					var err error
					status, err = executeCheck(nil)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				})

				if status != tt.wantStatus {
					t.Errorf("status = %d, want %d (output: %q)", status, tt.wantStatus, out)
				}
				if tt.wantStatus != sensu.CheckStateOK && !strings.Contains(out, "has been running for") {
					t.Errorf("output %q does not report the runtime", out)
				}
			})
		})
	}
}

func TestExecuteCheckMatchesCmdline(t *testing.T) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Skipf("cannot inspect own process: %v", err)
	}
	cmdline, err := p.Cmdline()
	if err != nil || cmdline == "" {
		t.Skipf("cannot determine own command line: %v", err)
	}

	const unreachable = 1e6

	// With --cmdline the plugin matches on the full command line, so the bare
	// process name must no longer match.
	cfg := Config{
		Process:    ownProcessName(t),
		CmdLine:    true,
		MemoryCrit: 0,
		MemoryWarn: 0,
		CPUCrit:    unreachable,
		CPUWarn:    unreachable,
	}
	withConfig(t, cfg, func() {
		captureOutput(t, func() {
			if status, _ := executeCheck(nil); status != sensu.CheckStateOK {
				t.Errorf("matching on the process name should not match a command line, got status %d", status)
			}
		})
	})

	cfg.Process = cmdline
	withConfig(t, cfg, func() {
		var status int
		out := captureOutput(t, func() {
			status, _ = executeCheck(nil)
		})
		if status != sensu.CheckStateCritical {
			t.Errorf("status = %d, want %d (critical) for command line %q", status, sensu.CheckStateCritical, cmdline)
		}
		if !strings.Contains(out, "% memory") {
			t.Errorf("output %q does not report memory usage", out)
		}
	})
}

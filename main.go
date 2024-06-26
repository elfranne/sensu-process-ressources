package main

import (
	"fmt"
	"math"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/shirou/gopsutil/v3/process"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	CPUWarn    float64
	CPUCrit    float64
	MemoryWarn float32
	MemoryCrit float32
	TimeWarn   int64
	TimeCrit   int64
	Scheme     string
	Process    string
	CmdLine    bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-process-ressources",
			Short:    "Check if process is using too much ressources (CPU/memory)",
			Keyspace: "sensu.io/plugins/check-process-ressources/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:     "process",
			Argument: "process",
			Default:  "",
			Usage:    "Process to monitor",
			Value:    &plugin.Process,
		},
		&sensu.PluginConfigOption[bool]{
			Path:     "cmdline",
			Argument: "cmdline",
			Default:  false,
			Usage:    "Use full command line of the process",
			Value:    &plugin.CmdLine,
		},
		&sensu.PluginConfigOption[float64]{
			Path:     "cpu-warn",
			Argument: "cpu-warn",
			Default:  float64(50),
			Usage:    "Warn if process is using more than cpu-warn (in percent)",
			Value:    &plugin.CPUWarn,
		},
		&sensu.PluginConfigOption[float64]{
			Path:     "cpu-crit",
			Argument: "cpu-crit",
			Default:  float64(75),
			Usage:    "Critical if process is using more than cpu-crit (in percent)",
			Value:    &plugin.CPUCrit,
		},
		&sensu.PluginConfigOption[float32]{
			Path:     "memory-warn",
			Argument: "memory-warn",
			Default:  float32(50),
			Usage:    "Warn if process is using more than memory-warn (in percent)",
			Value:    &plugin.MemoryWarn,
		},
		&sensu.PluginConfigOption[float32]{
			Path:     "memory-crit",
			Argument: "memory-crit",
			Default:  float32(70),
			Usage:    "Critical if process is using more than memory-crit (in percent)",
			Value:    &plugin.MemoryCrit,
		},
		&sensu.PluginConfigOption[int64]{
			Path:     "time-warn",
			Argument: "time-warn",
			Usage:    "Warn if process has been running for longer than time-warn (in seconds)",
			Value:    &plugin.TimeWarn,
		},
		&sensu.PluginConfigOption[int64]{
			Path:     "time-crit",
			Argument: "time-crit",
			Usage:    "Critical if process has been running for longer than time-crit (in seconds)",
			Value:    &plugin.TimeCrit,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	if plugin.CPUCrit == 100 {
		return sensu.CheckStateWarning, fmt.Errorf("that's just stupid")
	}
	if plugin.CPUWarn == 100 {
		return sensu.CheckStateWarning, fmt.Errorf("that's just stupid")
	}
	if plugin.MemoryCrit == 100 {
		return sensu.CheckStateWarning, fmt.Errorf("that's just stupid")
	}
	if plugin.MemoryWarn == 100 {
		return sensu.CheckStateWarning, fmt.Errorf("that's just stupid")
	}
	if plugin.Process == "" {
		return sensu.CheckStateWarning, fmt.Errorf("process is required")
	}

	return sensu.CheckStateOK, nil
}

func Round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func executeCheck(event *corev2.Event) (int, error) {
	process, _ := process.Processes()
	for _, p := range process {
		cpu, _ := p.CPUPercent()
		memory, _ := p.MemoryPercent()
		timems, _ := p.CreateTime() //create time is provide in millisecond epoch
		time := float64(time.Now().Unix()) - math.Round(float64(timems)/1000)
		name := ""
		if plugin.CmdLine {
			name, _ = p.Cmdline()
		} else {
			name, _ = p.Name()
		}

		// Warning memory
		if name == plugin.Process && memory >= plugin.MemoryWarn {
			fmt.Printf("%s is using %f %% memory, limit set at %f\n", plugin.Process, Round(float64(memory), 0.1), plugin.MemoryWarn)
			return sensu.CheckStateWarning, nil
		}
		// Warning CPU
		if name == plugin.Process && cpu >= plugin.CPUWarn {
			fmt.Printf("%s is using %f %% CPU, limit set at %f\n", plugin.Process, Round(float64(cpu), 0.1), plugin.CPUWarn)
			return sensu.CheckStateWarning, nil
		}
		// Critical memory
		if name == plugin.Process && memory >= plugin.MemoryCrit {
			fmt.Printf("%s is using %f %% memory, limit set at %f\n", plugin.Process, Round(float64(memory), 0.1), plugin.MemoryCrit)
			return sensu.CheckStateCritical, nil
		}
		// Critical CPU
		if name == plugin.Process && cpu >= plugin.CPUCrit {
			fmt.Printf("%s is using %f %% CPU, limit set at %f\n", plugin.Process, Round(float64(cpu), 0.1), plugin.CPUCrit)
			return sensu.CheckStateCritical, nil
		}
		// Warnning Time
		if name == plugin.Process && plugin.TimeWarn > 0 && time >= float64(plugin.TimeWarn) {
			fmt.Printf("%s has been running for %f seconds, limit set at %d\n", plugin.Process, time, plugin.TimeWarn)
			return sensu.CheckStateWarning, nil
		}
		// Critical Time
		if name == plugin.Process && plugin.TimeCrit > 0 && time >= float64(plugin.TimeCrit) {
			fmt.Printf("%s has been running for %f seconds, limit set at %d\n", plugin.Process, time, plugin.TimeCrit)
			return sensu.CheckStateCritical, nil
		}
	}
	return sensu.CheckStateOK, nil
}

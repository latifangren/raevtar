package handler

import "testing"

func TestPublicHostHealthFallsBackWhenTotalsAndTemperatureAreUnavailable(t *testing.T) {
	health := publicHostHealth(HostStats{})

	for name, got := range map[string]string{
		"cpu load":     health.CPULoad,
		"cpu cores":    health.CPUCores,
		"ram usage":    health.RAMUsage,
		"ram percent":  health.RAMPercent,
		"disk usage":   health.DiskUsage,
		"disk percent": health.DiskPercent,
		"temperature":  health.Temperature,
	} {
		if got != "N/A" {
			t.Fatalf("%s = %q, want N/A", name, got)
		}
	}
}

func TestPublicHostHealthFormatsValidZeroLoad(t *testing.T) {
	health := publicHostHealth(HostStats{
		CPU:           CPUStats{Load1: 0, Load5: 0.25, Load15: 0.5, Cores: 4},
		RAM:           RAMStats{Used: 1024 * 1024, Total: 2 * 1024 * 1024, Percent: 50},
		Disk:          DiskStats{Used: 8 * 1024 * 1024, Total: 32 * 1024 * 1024, Percent: 25},
		Temp:          48.5,
		TempAvailable: true,
	})

	for name, tt := range map[string]struct {
		got  string
		want string
	}{
		"cpu load":     {got: health.CPULoad, want: "0.00 / 0.25 / 0.50"},
		"cpu cores":    {got: health.CPUCores, want: "4"},
		"ram usage":    {got: health.RAMUsage, want: "1.0 / 2.0 GB"},
		"ram percent":  {got: health.RAMPercent, want: "50%"},
		"disk usage":   {got: health.DiskUsage, want: "8.0 / 32.0 GB"},
		"disk percent": {got: health.DiskPercent, want: "25%"},
		"temperature":  {got: health.Temperature, want: "48.5°C"},
	} {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("text = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

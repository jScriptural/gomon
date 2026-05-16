package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper around time.Duration to support JSON unmarshaling
// of duration strings like "500ms".
type Duration time.Duration

// Option holds the parsed commandline
// arguments
type Option map[string]string

type Hooks struct {
	PreBuild  string `json:"prebuild"`
	PostBuild string `json:"postbuild"`
	PreStart  string `json:"prestart"`
	PostStart string `json:"poststart"`
}

type Env map[string]string

type Config struct {
	Watch       []string `json:"watch"`
	Ext         []string `json:"ext"`
	Ignore      []string `json:"ignore"`
	Build       string   `json:"build"`
	Run         string   `json:"run"`
	Delay       Duration `json:"delay"`
	LegacyWatch bool     `json:"legacy_watch"`
	Hooks       Hooks    `json:"hooks"`
	Env         Env      `json:"env"`
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	case float64:
		*d = Duration(time.Duration(value) * time.Millisecond)
		return nil
	default:
		return fmt.Errorf("invalid duration type: %T", v)
	}
}

func (env Env) ToSlice() []string {
	var s []string
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}

	return s
}

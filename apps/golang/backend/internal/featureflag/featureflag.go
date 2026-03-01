package featureflag

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
)

// AllFlags lists every known flag key. Add new flags here.
var AllFlags []string

// Config holds the resolved flag values.
type Config struct {
	Flags map[string]bool
}

// LoadConfig reads FF_* environment variables and returns a Config.
// Missing or unparseable values default to true (features enabled).
func LoadConfig() Config {
	flags := make(map[string]bool, len(AllFlags))
	for _, f := range AllFlags {
		envKey := "FF_" + strings.ToUpper(f)
		v := strings.TrimSpace(os.Getenv(envKey))
		if v == "" {
			flags[f] = true
			continue
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			flags[f] = true
			continue
		}
		flags[f] = b
	}
	return Config{Flags: flags}
}

// Init registers the EnvProvider with the OpenFeature API.
func Init(cfg Config) {
	provider := newEnvProvider(cfg.Flags)
	if err := openfeature.SetProvider(provider); err != nil {
		log.Printf("featureflag: failed to set provider: %v", err)
	}
}

// IsEnabled evaluates a boolean feature flag via the OpenFeature client.
// Returns true if the flag is enabled, false otherwise.
func IsEnabled(flag string) bool {
	client := openfeature.NewClient("micro-dp")
	v, _ := client.BooleanValue(context.Background(), flag, false, openfeature.EvaluationContext{})
	return v
}

// LogStartup logs the current state of all feature flags.
func LogStartup(cfg Config) {
	var enabled, disabled []string
	for _, f := range AllFlags {
		if cfg.Flags[f] {
			enabled = append(enabled, f)
		} else {
			disabled = append(disabled, f)
		}
	}
	log.Printf("feature flags initialized provider=env enabled=[%s] disabled=[%s]",
		joinOrNone(enabled), joinOrNone(disabled))
}

func joinOrNone(ss []string) string {
	if len(ss) == 0 {
		return "none"
	}
	return fmt.Sprintf("%s", strings.Join(ss, ", "))
}

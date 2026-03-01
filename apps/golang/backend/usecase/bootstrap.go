package usecase

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/user/micro-dp/domain"
)

type BootstrapConfig struct {
	Enabled bool
	Emails  []string
}

func ParseBootstrapConfig(enabledStr, emailsCSV string) BootstrapConfig {
	enabled := strings.EqualFold(strings.TrimSpace(enabledStr), "true")
	var emails []string
	for _, e := range strings.Split(emailsCSV, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			emails = append(emails, e)
		}
	}
	return BootstrapConfig{Enabled: enabled, Emails: emails}
}

func BootstrapSuperadmins(ctx context.Context, users domain.UserRepository, cfg BootstrapConfig) error {
	if !cfg.Enabled {
		log.Println("bootstrap: superadmin bootstrap disabled")
		return nil
	}
	if len(cfg.Emails) == 0 {
		log.Println("bootstrap: SUPERADMIN_EMAILS is empty, skipping")
		return nil
	}

	for _, email := range cfg.Emails {
		user, err := users.FindByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				log.Printf("bootstrap: user %s not registered yet, skipping", email)
				continue
			}
			return err
		}

		if user.PlatformRole == domain.PlatformRoleSuperadmin {
			log.Printf("bootstrap: user %s already superadmin, skipping", email)
			continue
		}

		if err := users.UpdatePlatformRole(ctx, user.ID, domain.PlatformRoleSuperadmin); err != nil {
			return err
		}
		log.Printf("bootstrap: promoted user %s to superadmin", email)
	}

	return nil
}

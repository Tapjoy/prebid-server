package zeroclickfraud

import (
	"testing"
	"text/template"

	"github.com/prebid/prebid-server/privacy"
	"github.com/prebid/prebid-server/privacy/ccpa"
	"github.com/prebid/prebid-server/privacy/gdpr"
	"github.com/stretchr/testify/assert"
)

func TestZeroClickFraudSyncer(t *testing.T) {
	syncURL := "https://s.0cf.io/sync?gdpr={{.GDPR}}&gdpr_consent={{.GDPRConsent}}&us_privacy={{.USPrivacy}}&r=https%3A%2F%2Flocalhost%3A8888%2Fsetuid%3Fbidder%3Dzeroclickfraud%26gdpr%3D{{.GDPR}}%26gdpr_consent%3D{{.GDPRConsent}}%26uid%3D%24%7Buid%7D"
	syncURLTemplate := template.Must(
		template.New("sync-template").Parse(syncURL),
	)

	syncer := NewZeroClickFraudSyncer(syncURLTemplate)
	syncInfo, err := syncer.GetUsersyncInfo(privacy.Policies{
		GDPR: gdpr.Policy{
			Signal:  "1",
			Consent: "BONciguONcjGKADACHENAOLS1rAHDAFAAEAASABQAMwAeACEAFw",
		},
		CCPA: ccpa.Policy{
			Consent: "1NYN",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://s.0cf.io/sync?gdpr=1&gdpr_consent=BONciguONcjGKADACHENAOLS1rAHDAFAAEAASABQAMwAeACEAFw&us_privacy=1NYN&r=https%3A%2F%2Flocalhost%3A8888%2Fsetuid%3Fbidder%3Dzeroclickfraud%26gdpr%3D1%26gdpr_consent%3DBONciguONcjGKADACHENAOLS1rAHDAFAAEAASABQAMwAeACEAFw%26uid%3D%24%7Buid%7D", syncInfo.URL)
	assert.Equal(t, "iframe", syncInfo.Type)
	assert.Equal(t, false, syncInfo.SupportCORS)
}

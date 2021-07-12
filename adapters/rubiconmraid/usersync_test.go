package rubiconmraid

import (
	"testing"
	"text/template"

	"github.com/prebid/prebid-server/privacy"
	"github.com/prebid/prebid-server/privacy/gdpr"
	"github.com/stretchr/testify/assert"
)

func TestRubiconSyncer(t *testing.T) {
	syncURL := "https://pixel.rubiconproject.com/exchange/sync.php?p=prebid&gdpr={{.GDPR}}&gdpr_consent={{.GDPRConsent}}"
	syncURLTemplate := template.Must(
		template.New("sync-template").Parse(syncURL),
	)

	syncer := NewRubiconSyncer(syncURLTemplate)
	syncInfo, err := syncer.GetUsersyncInfo(privacy.Policies{
		GDPR: gdpr.Policy{
			Signal: "0",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://pixel.rubiconproject.com/exchange/sync.php?p=prebid&gdpr=0&gdpr_consent=", syncInfo.URL)
	assert.Equal(t, "redirect", syncInfo.Type)
	assert.Equal(t, false, syncInfo.SupportCORS)
	assert.Equal(t, "rubiconmraid", syncer.FamilyName())
}
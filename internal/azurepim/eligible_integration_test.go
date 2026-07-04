//go:build integration

package azurepim

import (
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func TestListForSubscription_liveIntegration(t *testing.T) {
	if os.Getenv("AZKIT_LIVE_TESTS") != "1" {
		t.Skip("set AZKIT_LIVE_TESTS=1 to run")
	}
	if os.Getenv("AZKIT_LIVE_SUBSCRIPTION") == "" {
		t.Skip("set AZKIT_LIVE_SUBSCRIPTION to a subscription ID")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("NewDefaultAzureCredential: %v", err)
	}

	source := azureEligibilitySchedules{cred: cred}
	as, err := source.ListForSubscription(t.Context(), os.Getenv("AZKIT_LIVE_SUBSCRIPTION"))
	if err != nil {
		t.Fatalf("ListForSubscription: %v", err)
	}
	t.Logf("got %d assignments", len(as))
}

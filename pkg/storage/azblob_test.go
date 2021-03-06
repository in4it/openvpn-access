package storage

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure/auth"
)

const onAzureVM = false

func TestAzConnectivitiy(t *testing.T) {
	if !onAzureVM {
		t.Skip("Not on azure VM")
	}
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		t.Errorf("Auth error: %s", err)
		return
	}

	msi := settings.GetMSI()
	msi.Resource = "https://storage.azure.com/"
	token, err := msi.ServicePrincipalToken()
	if err != nil {
		t.Errorf("Creds error: %s", err)
		return
	}

	if err := token.RefreshWithContext(context.Background()); err != nil {
		t.Errorf("refresh token error: %s", err)
		return
	}

	if len(token.OAuthToken()) == 0 {
		t.Errorf("Token is empty")
	}
	return

}

func TestAzBlobStorage(t *testing.T) {
	if os.Getenv("TEST_AZURE_STORAGE_ACCOUNT_NAME") == "" {
		t.Skip("env variables for test not set")
	}
	accountName := os.Getenv("TEST_AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("TEST_AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("TEST_AZURE_STORAGE_CONTAINER_NAME")
	azBlob, err := NewAzBlob(accountName, accountKey)
	if err != nil {
		t.Errorf("Error while doing NewAzBlob: %s", err)
	}

	// put object
	data := "test"

	err = azBlob.HeadObject(containerName, "test.txt")
	if err == nil {
		t.Errorf("Expected error, but didn't get error (headobject)")
	}

	err = azBlob.PutObject(containerName, "test.txt", data, "")
	if err != nil {
		t.Errorf("Error while doing PutObject: %s", err)
	}

	err = azBlob.HeadObject(containerName, "test.txt")
	if err != nil {
		t.Errorf("Error while doing HeadObject: %s", err)
	}

	out, err := azBlob.GetObject(containerName, "test.txt")
	if err != nil {
		t.Errorf("Error while doing GetObject: %s", err)
	}
	if out.String() != "test" {
		t.Errorf("Output is not expected (expected string), got: %s", out.String())
	}
	err = azBlob.DeleteObject(containerName, "test.txt")
	if err != nil {
		t.Errorf("Error while doing DeleteObject: %s", err)
	}
}

package storage

import (
	"os"
	"testing"
)

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

	err = azBlob.PutObject(containerName, "test.txt", data, "")
	if err != nil {
		t.Errorf("Error while doing PutObject: %s", err)
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

package src

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Returns the path to the downloaded image file.
func FetchNextDiskImage(ctx context.Context, lastImageVersion string, endpointUrl string, destPathPrefix string) (string, error) {
	newVersion, err := PollForNewImage(ctx, endpointUrl, 60, lastImageVersion)
	if err != nil {
		return "", err
	}
	log.Printf("New image version found: %s", newVersion)
	client := &http.Client{}
	destPath := fmt.Sprintf("%s_%s.vhd", destPathPrefix, newVersion)
	file, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", endpointUrl, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Downloaded new image version %s to %s", newVersion, destPath)
	return destPath, nil
}

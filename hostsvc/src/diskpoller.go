package src

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func fetchAndParse(ctx context.Context, client *http.Client, endpointUrl string) (Metadata, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpointUrl, nil)
	if err != nil {
		return Metadata{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Metadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var metadata Metadata
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return Metadata{}, err
	}
	return metadata, nil
}

// This function repeatedly polls the managementsvc endpoint for the latest metadata file about current image versions. Once a new image version is found, it returns the new version string. This function should be run in a goroutine that runs indefinitely until the host is shut down.
func PollForNewImage(ctx context.Context, endpointUrl string, periodSecs int, currVersion string) (string, error) {
	client := &http.Client{}
	for {
		metadata, err := fetchAndParse(ctx, client, endpointUrl)
		if err != nil {
			return "", err
		}
		if metadata.Next.Version != currVersion {
			return metadata.Next.Version, nil
		}
		log.Printf("No new image found. Current version: %s. Next version: %s. Polling again in %d seconds.", currVersion, metadata.Next.Version, periodSecs)
		// Sleep (cancellably) for the specified period before polling again
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(periodSecs) * time.Second):
		}
	}
}

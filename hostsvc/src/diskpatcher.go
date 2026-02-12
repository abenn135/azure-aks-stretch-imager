package src

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type LocalDiskPartitionMetadata struct {
	BootLabels []string `json:"bootLabels"`
}

func ApplyNewImage(ctx context.Context, basePath string, endpointUrl string) error {
	imagePathPrefix := fmt.Sprintf("%s/%s", basePath, time.Now().Format("2006-01-02_15-04-05"))
	imagePath, err := FetchNextDiskImage(ctx, "0.0.1", endpointUrl, imagePathPrefix)
	if err != nil {
		return err
	}

	log.Printf("Downloaded new image to %s", imagePath)

	// First, validate downloaded VHD and determine boot partition to copy to the target partition on the running node.
	c := os.exec.Command("/bin/bash", "-c", "virt-filesystems -a "+imagePath+" --partitions --lsblk --noheadings")
	output, err := c.Output()

	eligibleBootPartitionLabels := []string{}
	currentBootPartitionLabel := discoverCurrentBootPartition()
	partitionMetadataFilepath := fmt.Sprintf("%s/metadata.json", basePath)
	data, err := os.ReadFile(partitionMetadataFilepath)
	partitionMetadata := LocalDiskPartitionMetadata{}
	if err == nil {
		err = json.Unmarshal(data, &partitionMetadata)
		if err != nil {
			return err
		}
		eligibleBootPartitionLabels = partitionMetadata.BootLabels
	} else {
		if os.IsNotExist(err) {
			eligibleBootPartitionLabels = discoverLikelyBootPartitionLabels()
		} else {
			return err
		}
	}

	nextApplyPartitionLabel := selectNextBootPartition(currentBootPartitionLabel, eligibleBootPartitionLabels)

	return nil
}

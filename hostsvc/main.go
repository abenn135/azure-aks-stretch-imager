package main

import (
	"azure-aks-stretch-imager/hostsvc/src"
	"context"
)

func main() {
	/*
		Todo:

		1. start periodic polling goroutine that checks for new disk images.
		1. when a new image version is found, download it to the scratch partition.
		1. when a new image has been downloaded, extract it onto the non-running partition.
		1. Once that is done, update-grub and set a label on the local kubelet process to indicate that this node is ready to switch to the new image.
	*/
	ctx := context.Background()
	go src.FetchNextDiskImage(ctx, "0.0.1", "http://localhost:8080/metadata", "/tmp/next_image.img")

}

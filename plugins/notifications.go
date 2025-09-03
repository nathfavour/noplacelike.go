package plugins

import (
	// "fmt"
	"log"
	// "os"
	// "path/filepath"
	// "strings"
)

// basically, this module handles cross device notification sharing
//
//

type Notification struct {
	Message string
}

// SendNotification sends a notification to a particular device
func SendNotification(deviceID string, notification Notification) {
	// ensure all functionalities and errors are accounted for
	if deviceID == "" {
		log.Println("Device ID is empty")
	}

	// send notification
	if err := SendNotification(deviceID, notification); err != nil {
		log.Println("encountered the following error:", err)
	}

	// log successful notification
	log.Println("successfully sent notification to device!")
}

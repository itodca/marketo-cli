package cmd

import (
	"encoding/json"
	"fmt"
)

func folderValue(folderID int, folderType string) (string, error) {
	if folderID == 0 && folderType == "" {
		return "", nil
	}
	if folderID == 0 || folderType == "" {
		return "", fmt.Errorf("Both folder_id and folder_type are required together")
	}

	encoded, err := json.Marshal(map[string]any{"id": folderID, "type": folderType})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

package downloads

import (
	"fmt"
	"math"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/siku2/arigo"
)

type Status struct {
	Msg        *gotgbot.Message
	LastStatus string
}

type StatusMessage struct {
	message  string
	filename string
	filesize string
}

func GetStatus(details *DownloadDetails) (message string, filename string, filesizeStr string, err error) {
	status, err := aria2c.Aria.TellStatus(details.Gid, "status", "totalLength", "completedLength", "downloadSpeed", "files")
	if err != nil {
		return "", "", "", err
	} else if status.Status == arigo.StatusActive {
		statusMessage := generateStatusMessage(status.TotalLength, status.CompletedLength, status.DownloadSpeed, status.Files, false)
	}

}

func generateStatusMessage(totalSize uint, completed uint, speed uint, files []arigo.File, isUploading bool) StatusMessage {
	filePath := findAriaFilePath(files)
	fileName := getFileNameFromPath(filePath.Path, filePath.InputPath, filePath.DownloadUri)
	var progress float64 = 0
	if totalSize == 0 {
		progress = 0
	} else {
		progress = math.Round(float64(completed) * 100 / float64(totalSize))
	}

	
}

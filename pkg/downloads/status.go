package downloads

import (
	"fmt"
	"math"
	"sort"
	"time"

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

type StatusAll struct {
	message            string
	totalDownloadCount int
	singleStatuses     []*SingleStatus
}

func GetStatus(details *DownloadDetails) (message string, filename string, filesizeStr string, err error) {
	status, err := aria2c.Aria.TellStatus(details.Gid, "status", "totalLength", "completedLength", "downloadSpeed", "files")
	if err != nil {
		return "", "", "", err
	} else if status.Status == arigo.StatusActive {
		statusMessage := generateStatusMessage(status.TotalLength, status.CompletedLength, status.DownloadSpeed, status.Files, false)
		return statusMessage.message, statusMessage.filename, statusMessage.filesize, nil
	} else if details.IsUploading {
		var downloadSpeed int
		time := time.Now()

		if details.LastUploadCheckTimeStamp == nil {
			downloadSpeed = 0
		} else {
			downloadSpeed = (int(details.UploadedBytes) - int(details.UploadedBytesLast)) / int(time.Sub(*details.LastUploadCheckTimeStamp).Seconds())
		}

		details.UploadedBytesLast = details.UploadedBytes
		details.LastUploadCheckTimeStamp = &time

		statusMessage := generateStatusMessage(status.TotalLength, uint(details.UploadedBytes), uint(downloadSpeed), status.Files, true)
		return statusMessage.message, statusMessage.filename, statusMessage.filesize, nil
	} else {
		filePath := findAriaFilePath(status.Files)
		filename := getFileNameFromPath(filePath.Path, filePath.InputPath, filePath.DownloadUri)

		var message string
		if status.Status == arigo.StatusWaiting {
			message = fmt.Sprintf("<i>%s</i> - Queued", filename)
		} else {
			message = fmt.Sprintf("<i>%s</i> - %s", filename, status.Status)
		}
		return message, filename, "0B", nil
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

	totalLenStr := formatSize(int64(totalSize))
	progressString := generateProgress(progress)
	speedStr := formatSize(int64(speed))
	eta := downloadETA(totalSize, completed, speed)

	var Type string
	if isUploading {
		Type = "Uploading"
	} else {
		Type = "Filename"
	}

	message := fmt.Sprintf("<b>%s</b>: <code>%s</code>\n<b>Size</b>: <code>%s</code>\n<b>Progress</b>: <code>%s</code>\n<b>Speed</b>: <code>%sps</code>\n<b>ETA</b>: <code>%s</code>", Type, fileName, totalLenStr, progressString, speedStr, eta)
	return StatusMessage{
		message:  message,
		filename: fileName,
		filesize: totalLenStr,
	}
}

func downloadETA(totalLength uint, completedLength uint, speed uint) string {
	if speed == 0 {
		return "-"
	}
	time := (totalLength - completedLength) / speed
	seconds := time % 60
	minutes := (time / 60) % 60
	hours := float64(time / 3600)

	if hours == 0 {
		if minutes == 0 {
			return fmt.Sprintf("%vs", seconds)
		} else {
			return fmt.Sprintf("%vm %vs", minutes, seconds)
		}
	} else {
		return fmt.Sprintf("%vh %vm %vs", hours, minutes, seconds)
	}
}

type SingleStatus struct {
	Message  string
	Filename *string
	Details  *DownloadDetails
}

func getSingleStatus(details *DownloadDetails, msg *gotgbot.Message) *SingleStatus {
	var authCode int
	if msg != nil {
		authCode = IsAuthorized(msg, false)
	} else {
		authCode = 1
	}

	if authCode > -1 {
		msg, filename, _, err := GetStatus(details)
		if err != nil {
			return &SingleStatus{
				Message:  fmt.Sprintf("Error: %s - %s", details.Gid, err),
				Filename: nil,
				Details:  nil,
			}
		} else {
			return &SingleStatus{
				Message:  msg,
				Filename: &filename,
				Details:  details,
			}
		}
	} else {
		return &SingleStatus{Message: "You aren't authorized to use this bot here."}
	}
}

// implement sort.Interface
type ByTime []*SingleStatus

func (b ByTime) Len() int {
	return len(b)
}

func (b ByTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByTime) Less(i, j int) bool {
	return b[i].Details.StartTime.Before(b[j].Details.StartTime)
}

func GetStatusMessage() StatusAll {
	singleStatuses := make([]*SingleStatus, 0)
	GetDownloadManager().ForEachDownload(func(details *DownloadDetails) {
		singleStatuses = append(singleStatuses, getSingleStatus(details, nil))
	})

	if len(singleStatuses) > 0 {
		var message string
		sort.Sort(ByTime(singleStatuses))

		for i, ss := range singleStatuses {
			if i > 0 {
				message = fmt.Sprintf("%s\n\n%s", message, ss.Message)
			} else {
				message = ss.Message
			}
		}
		return StatusAll{
			message:            message,
			totalDownloadCount: len(singleStatuses),
			singleStatuses:     singleStatuses,
		}
	} else {
		return StatusAll{
			message:            "No active or queued downloads",
			totalDownloadCount: 0,
		}
	}
}

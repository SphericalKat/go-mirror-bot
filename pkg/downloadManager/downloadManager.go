package downloadmanager

import "github.com/PaulSonOfLars/gotgbot/v2"

var instance *DownloadManager = nil

type DownloadManager struct {

}

func (d *DownloadManager) AddDownload(gid string, dlDir string, msg *gotgbot.Message, isTar bool) {
	
}
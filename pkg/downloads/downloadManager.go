package downloads

import (
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/puzpuzpuz/xsync"
)

var instance *DownloadManager = nil

type DownloadManager struct {
	AllDownloads       *xsync.MapOf[*DownloadDetails]
	ActiveDownloads    *xsync.MapOf[*DownloadDetails]
	CancelledDownloads *xsync.MapOf[*DownloadDetails]
	CancelledMessages  *xsync.MapOf[[]string]
	AllStatuses        *xsync.MapOf[*Status]
	StatusQueue        *xsync.MPMCQueue
}

func GetDownloadManager() *DownloadManager {
	if instance == nil {
		instance = &DownloadManager{
			AllDownloads:       xsync.NewMapOf[*DownloadDetails](),
			ActiveDownloads:    xsync.NewMapOf[*DownloadDetails](),
			CancelledDownloads: xsync.NewMapOf[*DownloadDetails](),
			CancelledMessages:  xsync.NewMapOf[[]string](),
			AllStatuses:        xsync.NewMapOf[*Status](),
			StatusQueue:        xsync.NewMPMCQueue(1024),
		}
	}

	return instance
}

func (d *DownloadManager) AddDownload(gid string, dlDir string, msg *gotgbot.Message, isTar bool) {
	detail := NewDownloadDetails(gid, msg, isTar, dlDir)
	d.AllDownloads.Store(gid, detail)
}

func (d *DownloadManager) GetDownloadByGid(gid string) *DownloadDetails {
	detail, _ := d.AllDownloads.Load(gid)
	return detail
}

func (d *DownloadManager) DeleteDownload(gid string) {
	d.AllDownloads.Delete(gid)
	d.ActiveDownloads.Delete(gid)
}

// MoveDownloadToActive marks a download as active, once Aria2 starts downloading it.
func (d *DownloadManager) MoveDownloadToActive(detail *DownloadDetails) {
	detail.IsDownloading = true
	detail.IsUploading = false
	d.ActiveDownloads.Store(detail.Gid, detail)
}

// ChangeDownloadGid updates the GID of a download. This is needed if a download causes Aria2c to start
// another download, for example, in the case of BitTorrents. This function also
// marks the download as inactive, because we only find out about the new GID when
// Aria2c calls onDownloadComplete, at which point, the metadata download has been
// completed, but the files download hasn't yet started.
func (d *DownloadManager) ChangeDownloadGid(old, new string) {
	detail := d.GetDownloadByGid(old)
	d.DeleteDownload(old)
	detail.Gid = new
	detail.IsDownloading = false
	d.AllDownloads.Store(new, detail)
}

func (d *DownloadManager) GetDownloadByMsgId(msg *gotgbot.Message) *DownloadDetails {
	var detail *DownloadDetails
	d.AllDownloads.Range(func(key string, value *DownloadDetails) bool {
		if value.TgChatId == msg.Chat.Id && value.TgMsgId == msg.MessageId {
			detail = value
			return false
		}
		return true
	})
	return detail
}

func (d *DownloadManager) ForEachDownload(callback func(details *DownloadDetails)) {
	d.AllDownloads.Range(func(key string, value *DownloadDetails) bool {
		callback(value)
		return true
	})
}

func (d *DownloadManager) DeleteStatus(chatId int64) {
	d.AllStatuses.Delete(strconv.FormatInt(chatId, 10))
}

func (d *DownloadManager) GetStatus(chatId int64) *Status {
	val, _ := d.AllStatuses.Load(strconv.FormatInt(chatId, 10))
	return val
}

func (d *DownloadManager) AddStatus(msg *gotgbot.Message, lastStatus string) {
	d.AllStatuses.Store(strconv.FormatInt(msg.Chat.Id, 10), &Status{
		Msg:        msg,
		LastStatus: lastStatus,
	})
}

func (d *DownloadManager) ForEachStatus(callback func(status *Status)) {
	d.AllStatuses.Range(func(key string, value *Status) bool {
		callback(value)
		return true
	})
}

func (d *DownloadManager) QueueStatus(msg *gotgbot.Message, callback func(msg *gotgbot.Message, keep bool)) {
	d.StatusQueue.Enqueue(callback)
}

func (d *DownloadManager) AddCancelled(detail *DownloadDetails) {
	d.CancelledDownloads.Store(detail.Gid, detail)
	var message []string
	message, _ = d.CancelledMessages.Load(strconv.FormatInt(detail.TgChatId, 10))
	if message != nil {
		if isUnique(detail.TgUsername, message) {
			message = append(message, detail.TgUsername)
		}
	} else {
		message = []string{detail.TgUsername}
	}
	d.CancelledMessages.Store(strconv.FormatInt(detail.TgChatId, 10), message)
}

func (d *DownloadManager) ForEachCancelledDownload(callback func(detail *DownloadDetails)) {
	d.CancelledDownloads.Range(func(key string, value *DownloadDetails) bool {
		callback(value)
		return true
	})
}

func (d *DownloadManager) ForEachCancelledChat(callback func(chatId string, usernames []string)) {
	d.CancelledMessages.Range(func(key string, value []string) bool {
		callback(key, value)
		return true
	})
}

func (d *DownloadManager) RemoveCancelledMessage(chatId string) {
	d.CancelledMessages.Delete(chatId)
}

func (d *DownloadManager) RemoveCancelledDownload(gid string) {
	d.CancelledDownloads.Delete(gid)
}

func isUnique(toFind string, src []string) bool {
	for _, v := range src {
		if v == toFind {
			return false
		}
	}
	return true
}

package downloads

import (
	"os"
	"path/filepath"

	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/rs/zerolog/log"
)

const (
	PROGRESS_MAX_SIZE = 12
)

var PROGRESS_INCOMPLETE = []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}


func deleteDownloadedFile(subdir string) {
	path := filepath.Join(config.Conf.DownloadDir, subdir)
	err := os.RemoveAll(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("failed to delete download")
		return
	}
	log.Info().Str("path", path).Msg("deleted downloads")
}

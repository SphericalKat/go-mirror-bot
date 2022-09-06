package downloads

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

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

func formatSize(size int64) string {
	if size < 1000 {
		return fmt.Sprintf("%fB", formatNum(size))
	}

	if size < 1024000 {
		return fmt.Sprintf("%fKB", formatNum(size / 1024))
	}

	if size < 1048576000 {
		return fmt.Sprintf("%fMB", formatNum(size / 1048576))
	}

	return fmt.Sprintf("%fGB", formatNum(size / 1073741824))
}

func generateProgress(p int64) string {
	p = int64(math.Min(math.Max(float64(p), 0), 100))
	str := "["
	cFull := math.Floor(float64(p) / 8)
	cPart := p % 8 - 1
	str += strings.Repeat("█", int(cFull))

	if cPart >= 0 {
		str += PROGRESS_INCOMPLETE[cPart]
	}

	str += strings.Repeat(" ", PROGRESS_MAX_SIZE - int(cFull))
	str = fmt.Sprintf("%s] %d%", str, p)

	return str
}

func formatNum(n int64) float64 {
	return math.Round(float64(n) * 100) / 100
}
package downloads

import (
	"regexp"
	"strings"

	"github.com/SphericalKat/arigo"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
)

type FilePath struct {
	Path        string
	InputPath   string
	DownloadUri string
}

const TYPE_METADATA = "Metadata"

func findAriaFilePath(files []arigo.File) FilePath {
	filePath := files[0].Path
	var uri string
	if len(files[0].URIs) != 0 {
		uri = files[0].URIs[0].URI
	}

	if strings.Contains(filePath, config.Conf.DownloadDir) {
		split := strings.Split(filePath, ".")
		if split[len(split)-1] != "torrent" {
			// this is not a torrent's metadata
			return FilePath{
				Path:        filePath,
				InputPath:   filePath,
				DownloadUri: uri,
			}
		} else {
			return FilePath{
				Path:        "",
				InputPath:   filePath,
				DownloadUri: uri,
			}
		}
	} else {
		return FilePath{
			Path:        "",
			InputPath:   filePath,
			DownloadUri: uri,
		}
	}
}

func substr(s string, start, end int) string {
	counter, startIdx := 0, 0
	for i := range s {
		if counter == start {
			startIdx = i
		}
		if counter == end {
			return s[startIdx:i]
		}
		counter++
	}
	return s[startIdx:]
}

func getFileNameFromPath(filePath string, inputPath string, uri string) string {
	if filePath == "" {
		return getFileNameFromUri(inputPath, uri)
	}

	baseDirLength := len(config.Conf.DownloadDir) + 38
	nameEndIndex := strings.Index(filePath[baseDirLength:], "/")
	if nameEndIndex == -1 {
		nameEndIndex = len(filePath)
	}
	fileName := substr(filePath, baseDirLength, nameEndIndex)
	if fileName == "" {
		return getFileNameFromUri(inputPath, uri)
	}

	return fileName
}

func getFileNameFromUri(path string, uri string) string {
	if path != "" {
		if strings.HasPrefix(path, "[METADATA]") {
			return path[10:]
		} else {
			return TYPE_METADATA
		}
	} else {
		if uri != "" {
			re1 := regexp.MustCompile(`/#.*$|\/\?.*$|\?.*$/`)
			re2 := regexp.MustCompile(`/^.*\//`)
			uri = re1.ReplaceAllString(uri, "")
			uri = re2.ReplaceAllString(uri, "")
			return uri
		} else {
			return TYPE_METADATA
		}
	}
}

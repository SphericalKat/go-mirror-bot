package downloads

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	neturl "net/url"

	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/dghubble/trie"
	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
)

func ServiceAccount(credentialFile string) *http.Client {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error opening service account credentials")
	}
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			drive.DriveScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(context.TODO())
	return client
}

func UploadFile(details *DownloadDetails, filePath string, fileSize uint) {
	details.IsUploading = true
	fileName := getFileNameFromPath(filePath, "", "")
	realFilePath := getActualDownloadPath(filePath)
	DriveUploadFile(details, realFilePath, fileName, fileSize)

	// TODO: add tar support

}

func DriveUploadFile(details *DownloadDetails, filePath string, fileName string, fileSize uint) {
	log.Info().Str("path", filePath).Msg("Uploading to drive")

	client := ServiceAccount("service-account.json")

	srv, err := drive.New(client)
	if err != nil {
		log.Error().Err(err).Msg("Unable to create gdrive client")
	}

	// get parent dir
	dir := filepath.Dir(filePath)

	// construct new path trie
	pathTrie := trie.NewPathTrie()

	// create an fs representation from the base dir
	fileSys := os.DirFS(dir)

	hasFolder := false

	// walk the fs
	err = fs.WalkDir(fileSys, ".", func(path string, d fs.DirEntry, err error) error {
		// get information about the fs node
		info, er := d.Info()
		if er != nil {
			return er
		}

		// don't upload self, unnecessary nesting shenanigans occur otherwise
		if path == "." {
			return nil
		}

		// don't upload aria2 stuff
		if filepath.Ext(path) == ".aria2" {
			return fs.SkipDir
		}

		// check if the trie contains parent path. if yes, get the folder id and
		// use it as the parent id for the current fs node
		parent := filepath.Dir(path)
		parentId := config.Conf.DriveDirId
		if result := pathTrie.Get(parent); result != nil {
			parentId = result.(string)
		}

		// if node is a directory, create dir on google drive
		if info.IsDir() {
			hasFolder = true
			file, err := srv.Files.Create(&drive.File{
				Name:     info.Name(),
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{parentId},
			}).
				Fields("id").
				SupportsAllDrives(true).Do()
			if err != nil {
				return err
			}

			// put current path into trie for future children
			pathTrie.Put(path, file.Id)
		} else {
			absPath := filepath.Join(dir, path)
			// open a reader to file
			file, err := os.Open(absPath)
			if err != nil {
				return err
			}
			defer file.Close()

			// attempt to detect mime type
			var mimeType string
			m, err := mimetype.DetectFile(absPath)
			if err != nil {
				mimeType = "application/octet-stream"
			} else {
				mimeType = m.String()
			}

			_, err = srv.Files.
				Create(&drive.File{
					Name:     info.Name(),
					MimeType: mimeType,
					Parents:  []string{parentId},
				}).
				Media(file).
				ProgressUpdater(func(current, total int64) {
					details.UploadedBytes = current
				}).
				Fields("id").
				SupportsAllDrives(true).Do()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Unable to walk fs and upload file")
		cleanupDownload(details.Gid, fmt.Sprintf("Unable to upload to drive. %s", err), "", details)
	}

	fileSizeStr := formatSize(float64(fileSize))
	url := config.Conf.CloudflareIndex + "/" + neturl.PathEscape(fileName)
	if hasFolder {
		url += "/"
	}

	finalMessage := fmt.Sprintf("%s (%s)\n<a href=\"%s\">Cloudflare Link</a>", fileName, fileSizeStr, url)
	fmt.Println(finalMessage)
	cleanupDownload(details.Gid, finalMessage, "", details)
}

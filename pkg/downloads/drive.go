package downloads

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

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
	log.Info().Msg("Uploading file to drive " + filePath)
}
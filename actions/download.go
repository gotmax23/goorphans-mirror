package actions

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"go.gtmx.me/goorphans/common"
)

func Download(httpclient *http.Client, baseurl, dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
	for _, fn := range []string{common.OrphansJSON, common.OrphansTXT} {
		dlurl, err := url.JoinPath(baseurl, fn)
		if err != nil {
			return fmt.Errorf("failed to parse url: %v", err)
		}
		dlpath := path.Join(dir, fn)

		err = common.DownloadFile(httpclient, dlurl, dlpath)
		if err != nil {
			return err
		}

		// Make sure the downloaded data is valid
		if fn == common.OrphansJSON {
			_, err = common.LoadOrphans(dlpath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

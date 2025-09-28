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
	_, err := DownloadWithOrphans(httpclient, baseurl, dir)
	return err
}

// DownloadWithOrphans downloads data and return pared Orphans data
func DownloadWithOrphans(httpclient *http.Client, baseurl, dir string) (*common.Orphans, error) {
	var orphans *common.Orphans = nil
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return orphans, err
	}
	for _, fn := range []string{common.OrphansJSON, common.OrphansTXT} {
		dlurl, err := url.JoinPath(baseurl, fn)
		if err != nil {
			return orphans, fmt.Errorf("failed to parse url: %v", err)
		}
		dlpath := path.Join(dir, fn)

		err = common.DownloadFile(httpclient, dlurl, dlpath)
		if err != nil {
			return orphans, err
		}

		// Make sure the downloaded data is valid
		if fn == common.OrphansJSON {
			orphans, err = common.LoadOrphans(dlpath)
			if err != nil {
				return orphans, err
			}
		}
	}
	return orphans, err
}

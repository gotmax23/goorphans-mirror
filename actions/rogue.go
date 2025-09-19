package actions

import (
	"errors"
	"log"
	"net/http"

	mapset "github.com/deckarep/golang-set/v2"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
	"go.gtmx.me/goorphans/pagure"
)

func GetRoguePackagerGroupMembers(
	f *fasjson.EmailCacheClient,
	p *pagure.Client,
) (map[string][]string, error) {
	packagers, err := f.GetMembers("packager")
	packagerset := mapset.NewThreadUnsafeSet(packagers...)
	if err != nil {
		return nil, err
	}
	groups, err := p.GetGroups()
	if err != nil {
		return nil, err
	}
	r := make(map[string][]string, len(groups)+1)
	total := mapset.NewThreadUnsafeSet[string]()
	for _, group := range groups {
		if group == "packager" {
			continue
		}
		members, err := f.GetMembers(group)
		if err != nil {
			var serr *common.StatusCodeError
			if errors.As(err, &serr) && serr.StatusCode == http.StatusNotFound {
				log.Printf("skipping group %s: %v", group, serr)
				continue
			}
			return r, err
		}
		badmembers := []string{}
		for _, member := range members {
			if !packagerset.Contains(member) {
				badmembers = append(badmembers, member)
			}
		}
		if len(badmembers) > 0 {
			r[group] = badmembers
			total.Append(badmembers...)
		}

	}
	r["total"] = mapset.Sorted(total)
	return r, nil
}

// Package notifs sends notifcations to individual maintainers
package notifs

import (
	_ "embed"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/templates"
)

var UserTemplate = templates.Templates.Lookup("notifs_user.gotmpl")

type UserTemplateData struct {
	User     string
	Orphaned []string
	Indirect []string
}

const UserSubjectFmt = "Orphaned packages summary for @%s"

func GetUserTemplateData(o *common.Orphans, user string) *UserTemplateData {
	direct := o.AffectedPeople[user]
	slices.Sort(direct)
	directset := mapset.NewThreadUnsafeSet(direct...)
	all := o.AllAffectedPeople[user]
	allset := mapset.NewThreadUnsafeSet(all...)
	indirect := allset.Difference(directset)
	return &UserTemplateData{
		User:     user,
		Orphaned: direct,
		Indirect: mapset.Sorted(indirect),
	}
}

// TODO: pagure.io/fesco/issue/3475
// var FakeGroupAdminTemplate = templates.Templates.Lookup("notifs_fake-group-user.gotmpl")
//
// func GetFakeGroupAdminTemplateData(rogue *actions.RoguePackageAdmin, p *pagure.Client) {
// }

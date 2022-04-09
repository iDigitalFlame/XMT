//go:build crypt

package transform

import "github.com/iDigitalFlame/xmt/util/crypt"

func getDefultDomains() []string {
	var (
		r = make([]string, 0, 16)
		v = crypt.Get(241) // amazon.com\namazonaws.com\napple.com\naws.amazon.com\nbing.com\ndocs.google.com\nduckduckgo.com\nebay.com\nfacebook.com\ngithub.com\ngmail.com\ngoogle.com\nimages.google.com\nimg.t.co\ninstagram.com\nlinkedin.com\nlogin.live.com\nmaps.google.com\nmicrosoft.com\nmsn.com\noffice.com\noffice365.com\noutlook.com\noutlook.office.com\npaypal.com\nredd.it\nreddit.com\ns3.amazon.com\nsharepoint.com\nslack.com\nspotify.com\nt.co\ntwimg.com\ntwitch.tv\ntwitter.com\nupdate.windows.com\nwalmart.com\nwikipedia.org\nwindows.com\nxp.apple.com\nyahoo.com
	)
	for i, e := 0, 0; i < len(v); i++ {
		if v[i] != '\n' {
			continue
		}
		r = append(r, v[e:i])
		e = i + 1
	}
	return r
}

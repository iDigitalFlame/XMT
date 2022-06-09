// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

//go:build crypt

package transform

import "github.com/iDigitalFlame/xmt/util/crypt"

func getDefultDomains() []string {
	var (
		r = make([]string, 0, 16)
		v = crypt.Get(14) // amazon.com\namazonaws.com\napple.com\naws.amazon.com\nbing.com\ndocs.google.com\nduckduckgo.com\nebay.com\nfacebook.com\ngithub.com\ngmail.com\ngoogle.com\nimages.google.com\nimg.t.co\ninstagram.com\nlinkedin.com\nlogin.live.com\nmaps.google.com\nmicrosoft.com\nmsn.com\noffice.com\noffice365.com\noutlook.com\noutlook.office.com\npaypal.com\nredd.it\nreddit.com\ns3.amazon.com\nsharepoint.com\nslack.com\nspotify.com\nt.co\ntwimg.com\ntwitch.tv\ntwitter.com\nupdate.windows.com\nwalmart.com\nwikipedia.org\nwindows.com\nxp.apple.com\nyahoo.com
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

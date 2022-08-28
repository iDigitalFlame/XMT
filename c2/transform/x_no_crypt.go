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

//go:build !crypt

package transform

func getDefaultDomains() []string {
	return []string{
		"amazon.com",
		"amazonaws.com",
		"apple.com",
		"aws.amazon.com",
		"bing.com",
		"docs.google.com",
		"duckduckgo.com",
		"ebay.com",
		"facebook.com",
		"github.com",
		"gmail.com",
		"google.com",
		"images.google.com",
		"img.t.co",
		"instagram.com",
		"linkedin.com",
		"login.live.com",
		"maps.google.com",
		"microsoft.com",
		"msn.com",
		"office.com",
		"office365.com",
		"outlook.com",
		"outlook.office.com",
		"paypal.com",
		"redd.it",
		"reddit.com",
		"s3.amazon.com",
		"sharepoint.com",
		"slack.com",
		"spotify.com",
		"t.co",
		"twimg.com",
		"twitch.tv",
		"twitter.com",
		"update.windows.com",
		"walmart.com",
		"wikipedia.org",
		"windows.com",
		"xp.apple.com",
		"yahoo.com",
	}
}

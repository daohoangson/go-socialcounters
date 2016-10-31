package web

import (
	"regexp"

	"github.com/daohoangson/go-socialcounters/utils"
)

var basic *regexp.Regexp
var whitelistPrepared = false
var whitelist *regexp.Regexp
var blacklistPrepared = false
var blacklist *regexp.Regexp

func RulesAllowUrl(u utils.Utils, url string) bool {
	if !getBasic().MatchString(url) {
		u.Debugf("web.RulesAllowUrl: %s is not a valid url", url)
		return false
	}

	if wl := getWhitelist(u); wl != nil {
		if !wl.MatchString(url) {
			u.Debugf("web.RulesAllowUrl: %s is not whitelisted", url)
			return false
		}
	}

	if bl := getBlacklist(u); bl != nil {
		if bl.MatchString(url) {
			u.Debugf("web.RulesAllowUrl: %s is blacklisted", url)
			return false
		}
	}

	return true
}

func getBasic() *regexp.Regexp {
	if basic == nil {
		// this regex should filter out all weird strings
		basic = regexp.MustCompile(`^https?://[^\r\n]+$`)
	}

	return basic
}

func getWhitelist(u utils.Utils) *regexp.Regexp {
	if !whitelistPrepared {
		if value := u.ConfigGet("WHITELIST"); value != "" {
			whitelist, _ = regexp.Compile(value)
		}

		whitelistPrepared = true
	}

	return whitelist
}

func getBlacklist(u utils.Utils) *regexp.Regexp {
	if !blacklistPrepared {
		if value := u.ConfigGet("BLACKLIST"); value != "" {
			blacklist, _ = regexp.Compile(value)
		}

		blacklistPrepared = true
	}

	return blacklist
}

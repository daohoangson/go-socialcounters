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

	if wl := getWhitelist(u, false); wl != nil {
		if !wl.MatchString(url) {
			u.Debugf("web.RulesAllowUrl: %s is not whitelisted", url)
			return false
		}
	}

	if bl := getBlacklist(u, false); bl != nil {
		if bl.MatchString(url) {
			u.Debugf("web.RulesAllowUrl: %s is blacklisted", url)
			return false
		}
	}

	return true
}

func RulesRefresh(u utils.Utils) {
	getWhitelist(u, true)
	getBlacklist(u, true)

	u.Debugf("web.RulesReset ok")
}

func getBasic() *regexp.Regexp {
	if basic == nil {
		// this regex should filter out all weird strings
		basic = regexp.MustCompile(`^https?://[^\r\n]+$`)
	}

	return basic
}

func getWhitelist(u utils.Utils, refresh bool) *regexp.Regexp {
	if !whitelistPrepared || refresh {
		if value := u.ConfigGet("WHITELIST"); value != "" {
			compiled, err := regexp.Compile(value)
			if err != nil {
				u.Errorf("web.getWhitelist error on %s: %v", value, err)
			}

			whitelist = compiled
		}

		whitelistPrepared = true
	}

	return whitelist
}

func getBlacklist(u utils.Utils, refresh bool) *regexp.Regexp {
	if !blacklistPrepared  || refresh {
		if value := u.ConfigGet("BLACKLIST"); value != "" {
			compiled, err := regexp.Compile(value)
			if err != nil {
				u.Errorf("web.getBlacklist error on %s: %v", value, err)
			}

			blacklist = compiled
		}

		blacklistPrepared = true
	}

	return blacklist
}

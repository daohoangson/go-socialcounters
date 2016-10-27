package web

import (
	"os"
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

	if wl := getWhitelist(); wl != nil {
		if !wl.MatchString(url) {
			u.Debugf("web.RulesAllowUrl: %s is not whitelisted", url)
			return false
		}
	}

	if bl := getBlacklist(); bl != nil {
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

func getWhitelist() *regexp.Regexp {
	if !whitelistPrepared {
		if value := os.Getenv("WHITELIST"); value != "" {
			whitelist, _ = regexp.Compile(value)
		}

		whitelistPrepared = true
	}

	return whitelist
}

func getBlacklist() *regexp.Regexp {
	if !blacklistPrepared {
		if value := os.Getenv("BLACKLIST"); value != "" {
			blacklist, _ = regexp.Compile(value)
		}

		blacklistPrepared = true
	}

	return blacklist
}

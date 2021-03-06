package dnsfilter

import (
	"encoding/json"
	"net/http"

	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/urlfilter/rules"
)

var serviceRules map[string][]*rules.NetworkRule // service name -> filtering rules

type svc struct {
	name  string
	rules []string
}

// Keep in sync with:
// client/src/helpers/constants.js
// client/src/components/ui/Icons.js
var serviceRulesArray = []svc{
	{"whatsapp", []string{"||whatsapp.net^", "||whatsapp.com^"}},
	{"facebook", []string{"||facebook.com^", "||facebook.net^", "||fbcdn.net^", "||fb.me^", "||fb.com^", "||fbsbx.com^", "||messenger.com^"}},
	{"twitter", []string{"||twitter.com^", "||t.co^", "||twimg.com^"}},
	{"youtube", []string{"||youtube.com^", "||ytimg.com^", "||youtu.be^", "||googlevideo.com^", "||youtubei.googleapis.com^"}},
	{"twitch", []string{"||twitch.tv^", "||ttvnw.net^"}},
	{"netflix", []string{"||nflxext.com^", "||netflix.com^"}},
	{"instagram", []string{"||instagram.com^", "||cdninstagram.com^"}},
	{"snapchat", []string{"||snapchat.com^", "||sc-cdn.net^", "||impala-media-production.s3.amazonaws.com^"}},
	{"discord", []string{"||discord.gg^", "||discordapp.net^", "||discordapp.com^"}},
	{"ok", []string{"||ok.ru^"}},
	{"skype", []string{"||skype.com^"}},
	{"vk", []string{"||vk.com^"}},
	{"origin", []string{"||origin.com^", "||signin.ea.com^", "||accounts.ea.com^"}},
	{"steam", []string{"||steam.com^"}},
	{"epic_games", []string{"||epicgames.com^"}},
	{"reddit", []string{"||reddit.com^", "||redditstatic.com^", "||redditmedia.com^", "||redd.it^"}},
	{"mail_ru", []string{"||mail.ru^"}},
	{"cloudflare", []string{
		"||cloudflare.com^",
		"||cloudflare-dns.com^",
		"||cloudflare.net^",
		"||cloudflareinsights.com^",
		"||cloudflarestream.com^",
		"||cloudflareresolve.com^",
		"||cloudflareclient.com^",
		"||cloudflarebolt.com^",
		"||cloudflarestatus.com^",
		"||cloudflare.cn^",
		"||one.one^",
		"||warp.plus^",
	}},
	{"amazon", []string{
		"||amazon.com^",
		"||media-amazon.com^",
		"||images-amazon.com^",
		"||a2z.com^",
		"||amazon.ae^",
		"||amazon.ca^",
		"||amazon.cn^",
		"||amazon.de^",
		"||amazon.es^",
		"||amazon.fr^",
		"||amazon.in^",
		"||amazon.it^",
		"||amazon.nl^",
		"||amazon.com.au^",
		"||amazon.com.br^",
		"||amazon.co.jp^",
		"||amazon.com.mx^",
		"||amazon.co.uk^",
	}},
	{"ebay", []string{
		"||ebay.com^",
		"||ebayimg.com^",
		"||ebaystatic.com^",
		"||ebaycdn.net^",
		"||ebayinc.com^",
		"||ebay.at^",
		"||ebay.be^",
		"||ebay.ca^",
		"||ebay.ch^",
		"||ebay.cn^",
		"||ebay.de^",
		"||ebay.es^",
		"||ebay.fr^",
		"||ebay.ie^",
		"||ebay.in^",
		"||ebay.it^",
		"||ebay.ph^",
		"||ebay.pl^",
		"||ebay.nl^",
		"||ebay.com.au^",
		"||ebay.com.cn^",
		"||ebay.com.hk^",
		"||ebay.com.my^",
		"||ebay.com.sg^",
		"||ebay.co.uk^",
	}},
	{"tiktok", []string{
		"||tiktok.com^",
		"||tiktokcdn.com^",
		"||snssdk.com^",
		"||amemv.com^",
		"||toutiao.com^",
		"||ixigua.com^",
		"||pstatp.com^",
		"||ixiguavideo.com^",
		"||toutiaocloud.com^",
		"||toutiaocloud.net^",
		"||bdurl.com^",
		"||bytecdn.cn^",
		"||byteimg.com^",
		"||ixigua.com^",
		"||muscdn.com^",
		"||bytedance.map.fastly.net^",
	}},
}

// convert array to map
func initBlockedServices() {
	serviceRules = make(map[string][]*rules.NetworkRule)
	for _, s := range serviceRulesArray {
		netRules := []*rules.NetworkRule{}
		for _, text := range s.rules {
			rule, err := rules.NewNetworkRule(text, 0)
			if err != nil {
				log.Error("rules.NewNetworkRule: %s  rule: %s", err, text)
				continue
			}
			netRules = append(netRules, rule)
		}
		serviceRules[s.name] = netRules
	}
}

// BlockedSvcKnown - return TRUE if a blocked service name is known
func BlockedSvcKnown(s string) bool {
	_, ok := serviceRules[s]
	return ok
}

// ApplyBlockedServices - set blocked services settings for this DNS request
func (d *Dnsfilter) ApplyBlockedServices(setts *RequestFilteringSettings, list []string, global bool) {
	setts.ServicesRules = []ServiceEntry{}
	if global {
		d.confLock.RLock()
		defer d.confLock.RUnlock()
		list = d.Config.BlockedServices
	}
	for _, name := range list {
		rules, ok := serviceRules[name]

		if !ok {
			log.Error("unknown service name: %s", name)
			continue
		}

		s := ServiceEntry{}
		s.Name = name
		s.Rules = rules
		setts.ServicesRules = append(setts.ServicesRules, s)
	}
}

func (d *Dnsfilter) handleBlockedServicesList(w http.ResponseWriter, r *http.Request) {
	d.confLock.RLock()
	list := d.Config.BlockedServices
	d.confLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(list)
	if err != nil {
		httpError(r, w, http.StatusInternalServerError, "json.Encode: %s", err)
		return
	}
}

func (d *Dnsfilter) handleBlockedServicesSet(w http.ResponseWriter, r *http.Request) {
	list := []string{}
	err := json.NewDecoder(r.Body).Decode(&list)
	if err != nil {
		httpError(r, w, http.StatusBadRequest, "json.Decode: %s", err)
		return
	}

	d.confLock.Lock()
	d.Config.BlockedServices = list
	d.confLock.Unlock()

	log.Debug("Updated blocked services list: %d", len(list))

	d.ConfigModified()
}

// registerBlockedServicesHandlers - register HTTP handlers
func (d *Dnsfilter) registerBlockedServicesHandlers() {
	d.Config.HTTPRegister("GET", "/control/blocked_services/list", d.handleBlockedServicesList)
	d.Config.HTTPRegister("POST", "/control/blocked_services/set", d.handleBlockedServicesSet)
}

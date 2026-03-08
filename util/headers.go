package util

import (
	"fmt"
	"time"

	"github.com/corpix/uarand"

	"telkomsel-bot/config"
)

func GenerateHash() string {
	return RandomHex(28)
}

func GenerateTransactionID() string {
	now := time.Now()
	return fmt.Sprintf("A%s148700", now.Format("060102150405000000"))
}

func RandomUA() string {
	return uarand.GetRandom()
}

func BuildHeaders(accessAuth, authorization, msisdn, xDevice, webAppVersion string) map[string][]string {
	if webAppVersion == "" {
		webAppVersion = config.WebAppVersion
	}

	return map[string][]string{
		"accept":                      {"application/json"},
		"accept-encoding":             {"gzip, deflate, br, zstd"},
		"accept-language":             {"id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7"},
		"accessauthorization":         {fmt.Sprintf("Bearer %s", accessAuth)},
		"authorization":               {fmt.Sprintf("Bearer %s", authorization)},
		"authserver":                  {"2"},
		"channelid":                   {"WEB"},
		"content-type":                {"application/json"},
		"dnt":                         {"1"},
		"hash":                        {GenerateHash()},
		"language":                    {"id"},
		"mytelkomsel-web-app-version": {webAppVersion},
		"origin":                      {"https://my.telkomsel.com"},
		"priority":                    {"u=1, i"},
		"referer":                     {"https://my.telkomsel.com/"},
		"sec-ch-ua":                   {`"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`},
		"sec-ch-ua-mobile":            {"?0"},
		"sec-ch-ua-platform":          {`"Windows"`},
		"sec-fetch-dest":              {"empty"},
		"sec-fetch-mode":              {"cors"},
		"sec-fetch-site":              {"same-site"},
		"transactionid":               {GenerateTransactionID()},
		"user-agent":                  {RandomUA()},
		"web-msisdn":                  {msisdn},
		"x-device":                    {xDevice},
	}
}

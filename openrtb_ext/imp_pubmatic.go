package openrtb_ext

import "encoding/json"

// ExtImpPubmatic defines the contract for bidrequest.imp[i].ext.pubmatic
type ExtImpPubmatic struct {
	AdSlot   string                  `json:"adSlot"`
	PubID    string                  `json:"publisherId"`
	WrapExt  json.RawMessage         `json:"wrapper"`
	Keywords []*ImpExtPubmaticKeyVal `json:"keywords"`
	Dctr     json.RawMessage         `json:"dctr"`
}

// ImpExtPubmaticKeyVal defines the contract for bidrequest.imp[i].ext.pubmatic.keywords[i]
type ImpExtPubmaticKeyVal struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"vals,omitempty"`
}

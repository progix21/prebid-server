package pubmatic

import (
	"testing"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// TestValidParamsForPubmatic makes sure that the pubmatic schema accepts all imp.ext fields provided in  bidder-params
func TestValidParamsForPubmatic(t *testing.T) {
	validator, err := openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
	if err != nil {
		t.Fatalf("Failed to fetch the json-schemas. %v", err)
	}

	for _, validParam := range validParams {
		if err := validator.Validate(openrtb_ext.BidderPubmatic, openrtb.RawJSON(validParam)); err != nil {
			t.Errorf("Incorrect pubmatic params: %s", validParam)
		}
	}
}

// TestInvalidParamsForPubmatic  sure that the pubmatic schema rejects all imp.ext fields which are not consitant with provided in  bidder-params
func TestInvalidParamsForPubmatic(t *testing.T) {
	validator, err := openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
	if err != nil {
		t.Fatalf("Failed to fetch the json-schemas. %v", err)
	}

	for _, invalidParam := range invalidParams {
		if err := validator.Validate(openrtb_ext.BidderPubmatic, openrtb.RawJSON(invalidParam)); err == nil {
			t.Errorf("Schema allowed unexpected params: %s", invalidParam)
		}
	}
}

var validParams = []string{
	`{"adSlot":"AdTag_Div1@728x90:0","publisherId":"5890"}`,
	`{"adSlot":"AdTag_Div1@728x90:0","publisherId":"5890","wrapper":{"version":2,"profile":595}}`,
	`{"adSlot":"AdTag_Div1@728x90:0","publisherId":"5890","wrapper":{"version":2,"profile":595},"keywords":[{"key":"Key_1","vals":["Val_1","Val_2"]},{"key":"Key_2","vals":["Val_1","Val_2"]}]}`,
}

var invalidParams = []string{
	``,
	`null`,
	`{}`,
	`{"adSlot":"AdTag_Div1@728x90:0"}`,
	`{"publisherId":"5890"}`,
}

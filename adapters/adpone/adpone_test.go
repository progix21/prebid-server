package adpone

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters/adapterstest"
)

const testsDir = "adponetest"
const testsBidderEndpoint = "http://localhost/prebid_server"

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, testsDir, NewAdponeBidder(testsBidderEndpoint))
}

package ctv

import (
	"bytes"
	"fmt"
)

//ImpBidsIndex will return index of ad in each impression
type ImpBidsIndex []int

//Combinations get combinations of n ImpBidsIndexs
type Combinations struct {
	IAdPodGenerator
	Ads        []ImpBids //used to store all ads
	cache      ICache    //used to store result of comparision of ads
	comparator Comparator
	state      struct {
		state     ImpBidsIndex
		adpod     ImpBids
		lastIndex int
	}
}

//NewCombinations returns object of combinations
func NewCombinations(ads []ImpBids, comp Comparator) *Combinations {
	obj := &Combinations{
		Ads:        ads,
		cache:      NewMapCache(),
		comparator: comp,
	}
	return obj
}

//getNextAdPodCombination will return next valid combination
func (c *Combinations) getNextAdPodCombination() ImpBidsIndex {
	if nil == c.state.state {
		c.state.state = make(ImpBidsIndex, len(c.Ads))
		c.state.adpod = make(ImpBids, len(c.Ads))
	} else {
		for c.state.lastIndex >= 0 {
			if c.state.state[c.state.lastIndex] == len(c.Ads[c.state.lastIndex])-1 {
				c.state.lastIndex--
				//continue
			}
			c.state.state[c.state.lastIndex]++
			for j := c.state.lastIndex + 1; j < len(c.Ads); j++ {
				c.state.state[j] = 0
			}
			break
		}
		if c.state.lastIndex < 0 {
			return nil
		}
	}
	c.state.lastIndex = len(c.Ads) - 1
	return c.state.state[:]
}

//validateAdPod will validate adpod combination and return status
func (c *Combinations) validateAdPod(comb ImpBidsIndex) (int, int, bool) {
	//We can cache combinations
	for step := 1; step < len(comb); step++ {
		for i := 0; i+step < len(comb); i++ {
			next := i + step
			//fmt.Printf("(%d,%d)\n", c.ImpBidsIndexs[i][comb[i]], c.ImpBidsIndexs[next][comb[next]])

			value, present := c.cache.Get(i, comb[i], next, comb[next])
			if !present {
				value = c.comparator(c.Ads[i][comb[i]], c.Ads[next][comb[next]])
				c.cache.Set(i, comb[i], next, comb[next], value)
			}
			if false == value {
				//combination failed
				return i, next, false
			}
		}
	}
	return 0, 0, true
}

//GetAdPod will return next valid combination
func (c *Combinations) GetAdPod() ImpBids {
	result := c.getNextAdPodCombination()
	for nil != result {
		_, j, valid := c.validateAdPod(result[:])
		if valid {
			break
		}
		c.state.lastIndex = j
		result = c.getNextAdPodCombination()
	}

	if result == nil {
		return nil
	}

	for i, val := range result {
		c.state.adpod[i] = c.Ads[i][val]
	}
	return c.state.adpod[:]
}

func (c *Combinations) printCombination(adpod ImpBids) string {
	var buff bytes.Buffer
	buff.WriteByte('[')
	for i, ad := range adpod {
		if i != 0 {
			buff.WriteString(", ")
		}
		buff.WriteString(fmt.Sprint(ad.ImpID))
	}
	buff.WriteByte(']')
	return buff.String()
}

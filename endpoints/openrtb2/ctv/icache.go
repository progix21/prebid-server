package ctv

//ICache Interface to store cached  combinations
type ICache interface {
	Set(xImp, xIndex, yImp, yIndex int, value bool)
	Get(xImp, xIndex, yImp, yIndex int) (bool, bool)
}

package main

import (
	"math"

	"github.com/pkg/errors"
)

const nPages = 3     // page 1 (rank 1-30), page 2, ...
const nPageTypes = 5 // new, top, etc

type pageCoefficients = struct {
	pageTypeCoefficient float64
	pageCoefficient     float64
	rankCoefficients    [nPages]float64
}

// These coefficients are the output of regression-pagetype-page-and-rank.R
// from the hacker-news-data repository. The rank coefficients for the
// ask and best pages looked pretty questionable, due to a sparsity of data,
// so I manually estimated some values
var intercept = -2.31114
var coefficients = [nPageTypes]pageCoefficients{
	{0.0, -3.54835, [nPages]float64{-0.64681, -0.36647, -0.19429}},
	{-3.05791, -2.45658, [nPages]float64{-0.44322, -0.29721, -0.11946}},
	{-4.25609, -1.84011, [nPages]float64{-0.38478, -0.16904, 0.13236}},
	{-2.46316, -5.67281, [nPages]float64{-1.25747, -0.58459, -0.09087}},
	{-5.40623, -4.08824, [nPages]float64{-0.64128, -0.21784, 0.85713}},
}

func expectedUpvoteShare(pageType, oneBasedRank int) float64 {
	zeroBasedPage := (oneBasedRank - 1) / 30
	oneBasedRankOnPage := ((oneBasedRank - 1) % 30) + 1

	cs := coefficients[pageType]

	logAttentionShare := intercept + cs.pageTypeCoefficient +
		cs.pageCoefficient*math.Log(float64(zeroBasedPage+1)) +
		cs.rankCoefficients[zeroBasedPage]*math.Log(float64(oneBasedRankOnPage))

	return math.Exp(logAttentionShare)
}


func accumulateAttention(ndb newsDatabase, logger leveledLogger, pageType int, storyID int, oneBasedRank int, sampleTime int64, deltaUpvotes int, sitewideUpvotes int) [2]float64 {

	attentionShare := expectedUpvoteShare(pageType, oneBasedRank)
	deltaAttention := float64(sitewideUpvotes) * attentionShare

	// logger.Debug(
	// 	"Updating cumulative attention",
	// 	"pageType", pageType,
	// 	"oneBasedPage", zeroBasedPage+1,
	// 	"oneBasedRankOnPage", oneBasedRankOnPage,
	// 	"deltaUpvotes", deltaUpvotes,
	// 	"deltaAttention", deltaAttention,
	// 	"sitewideUpvotes", sitewideUpvotes,
	// 	"pageTypeCoefficient", cs.pageTypeCoefficient,
	// 	"term2", cs.pageCoefficient*math.Log(float64(zeroBasedPage+1)),
	// 	"term3", cs.rankCoefficients[zeroBasedPage]*math.Log(float64(oneBasedRankOnPage)),
	// 	"logAttentionShare", logAttentionShare,
	// 	"attentionShare", math.Exp(logAttentionShare))

	err := ndb.upsertAttention(storyID, deltaUpvotes, deltaAttention, sampleTime)
	if err != nil {
		logger.Err(errors.Wrap(err, "upsertAttention"))
		return [2]float64{}
	}
	return [2]float64{deltaAttention, attentionShare}
}

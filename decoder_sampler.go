package golsv

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Decoder interface {
	Decode(syndrome BinaryVector) (err error, errorVec BinaryVector)
	Length() int
	Name() string
	SameCoset(e, c BinaryVector) bool
	Syndrome(error BinaryVector) BinaryVector
}

type DecoderSampler struct {
	decoder Decoder
	errorWeight int
	samplesPerWeight int
	results DecoderSamplerResults
	resultsFilename string
	verbose bool
	failFast bool
}

type DecoderSamplerResults struct {
	DecoderType string
	ErrorWeight int
	SuccessCount int
	FailCount int
	EqualCount int
	SameCosetCount int
}

func NewDecoderSampler(decoder Decoder, errorWeight int, samplesPerWeight int, resultsFilename string, verbose bool, failFast bool) *DecoderSampler {
	return &DecoderSampler{
		decoder: decoder,
		errorWeight: errorWeight,
		samplesPerWeight: samplesPerWeight,
		resultsFilename: resultsFilename,
		verbose: verbose,
		failFast: failFast,
	}
}

func (S *DecoderSampler) Run() {
	n := S.decoder.Length()
	log.Printf("Sampling error weight=%d samples=%d", S.errorWeight, S.samplesPerWeight)
	S.results.DecoderType = S.decoder.Name()
	S.results.ErrorWeight = S.errorWeight
	S.results.SuccessCount = 0
	S.results.EqualCount = 0
	S.results.SameCosetCount = 0
	S.results.FailCount = 0
	for i := 0; i < S.samplesPerWeight; i++ {
		errorVec := NewBinaryVector(n)
		errorVec.RandomizeWithWeight(S.errorWeight)
		syndrome := S.decoder.Syndrome(errorVec)
		before := time.Now()
		err, decodedErrorVec := S.decoder.Decode(syndrome)
		status := "failure"
		now := time.Now()
		elapsed := now.Sub(before)
		if err != nil {
			S.results.FailCount++
		} else if decodedErrorVec.Equal(errorVec) {
			status = "success"
			S.results.SuccessCount++
			S.results.EqualCount++
		} else if S.decoder.SameCoset(errorVec, decodedErrorVec) {
			status = "success"
			S.results.SuccessCount++
			S.results.SameCosetCount++
		} else {
			S.results.FailCount++
		}
		if S.verbose {
			log.Printf("Decode %s: errweight=%d equal=%d sameCoset=%d fail=%d success=%d/%d/%d successrate=%1.1f%% dur=%d",
				status, S.errorWeight, S.results.EqualCount, S.results.SameCosetCount, S.results.FailCount, S.results.SuccessCount, i+1, S.samplesPerWeight,
				float64(S.results.SuccessCount*100)/float64(i+1), int(elapsed.Seconds()))
		}
		S.writeResultsFile()
		if S.failFast && S.results.FailCount > 0 {
			if S.verbose {
				log.Printf("FailFast mode: stopping after first failure at sample %d/%d", i+1, S.samplesPerWeight)
			}
			break
		}
	}
}

func (S *DecoderSampler) writeResultsFile() {
	if S.resultsFilename == "" {
		return
	}
	jsonData, err := json.MarshalIndent(S.results, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(S.resultsFilename, jsonData, 0644)
	if err != nil {
		panic(err)
	}
	log.Printf("Wrote file %s", S.resultsFilename)
}

type DecoderThresholdFinder struct {
	decoder Decoder
	minErrorWeight int
	maxErrorWeight int
	samplesPerWeight int
	verbose bool
}

type DecoderThresholdResults struct {
	ThresholdWeight int
	MaxSuccessWeight int
	MinFailureWeight int
	TotalSamples int
	SearchSteps int
}

func NewDecoderThresholdFinder(decoder Decoder, minErrorWeight int, maxErrorWeight int, samplesPerWeight int, verbose bool) *DecoderThresholdFinder {
	return &DecoderThresholdFinder{
		decoder: decoder,
		minErrorWeight: minErrorWeight,
		maxErrorWeight: maxErrorWeight,
		samplesPerWeight: samplesPerWeight,
		verbose: verbose,
	}
}

func (F *DecoderThresholdFinder) FindThreshold() DecoderThresholdResults {
	results := DecoderThresholdResults{
		ThresholdWeight: -1,
		MaxSuccessWeight: -1,
		MinFailureWeight: -1,
		TotalSamples: 0,
		SearchSteps: 0,
	}
	
	maxSuccessWeight := -1
	minFailureWeight := -1
	
	// Start with minimum error weight and double until we find failures
	currentWeight := F.minErrorWeight
	
	log.Printf("Starting doubling search for decoder threshold from weight %d", currentWeight)
	
	for currentWeight <= F.maxErrorWeight {
		results.SearchSteps++
		
		if F.verbose {
			log.Printf("Doubling search step %d: testing weight %d", results.SearchSteps, currentWeight)
		}
		
		sampler := NewDecoderSampler(F.decoder, currentWeight, F.samplesPerWeight, "", F.verbose, true)
		sampler.Run()
		actualSamples := sampler.results.SuccessCount + sampler.results.FailCount
		results.TotalSamples += actualSamples
		
		successRate := float64(sampler.results.SuccessCount) / float64(actualSamples)
		
		if F.verbose {
			log.Printf("Weight %d: success rate = %1.2f%% (%d/%d)", currentWeight, successRate*100, sampler.results.SuccessCount, actualSamples)
		}
		
		if sampler.results.FailCount == 0 {
			// All samples succeeded, continue doubling
			maxSuccessWeight = currentWeight
			// Special case: go from 0 to 1, then double normally
			if currentWeight == 0 {
				currentWeight = 1
			} else {
				currentWeight *= 2
			}
			// Don't exceed the maximum weight
			if currentWeight > F.maxErrorWeight {
				currentWeight = F.maxErrorWeight
			}
		} else {
			// Found failures, this is our first failure point
			minFailureWeight = currentWeight
			break
		}
	}
	
	results.MaxSuccessWeight = maxSuccessWeight
	results.MinFailureWeight = minFailureWeight
	
	if maxSuccessWeight != -1 && minFailureWeight != -1 {
		results.ThresholdWeight = maxSuccessWeight
	} else if minFailureWeight != -1 {
		results.ThresholdWeight = minFailureWeight
	}
	
	log.Printf("Threshold search completed: maxSuccessWeight=%d, minFailureWeight=%d, threshold=%d, steps=%d, totalSamples=%d", 
		results.MaxSuccessWeight, results.MinFailureWeight, results.ThresholdWeight, results.SearchSteps, results.TotalSamples)
	
	return results
}
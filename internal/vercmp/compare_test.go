package vercmp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/vercmp"
)

func TestTheBiggerNumberComparesHigher(t *testing.T) {
	// act
	compared := vercmp.Compare("9", "10")

	// assert
	assert.Negative(t, compared)
}

func TestLeadingZerosOfANumberDoNotCount(t *testing.T) {
	// act
	compared := vercmp.Compare("010", "10")

	// assert
	assert.Zero(t, compared)
}

func TestNumbersBetweenDotsCompareOneByOne(t *testing.T) {
	// act
	compared := vercmp.Compare("1.2.1", "1.10")

	// assert
	assert.Negative(t, compared)
}

func TestCharactersOutsideTheFormatAreSkipped(t *testing.T) {
	// act
	compared := vercmp.Compare("1+", "1")

	// assert
	assert.Zero(t, compared)
}

func TestAMinusMakesTheOlderVersion(t *testing.T) {
	// act
	compared := vercmp.Compare("123-1", "123.1")
	reversed := vercmp.Compare("123.1", "123-1")

	// assert
	assert.Negative(t, compared)
	assert.Positive(t, reversed)
}

func TestWhatFollowsAMinusOnBothSidesDecides(t *testing.T) {
	// act
	compared := vercmp.Compare("123-2", "123-1")

	// assert
	assert.Positive(t, compared)
}

func TestACaretMakesTheOlderVersion(t *testing.T) {
	// act
	compared := vercmp.Compare("123^post1", "123.a-1")
	reversed := vercmp.Compare("123.a-1", "123^post1")

	// assert
	assert.Negative(t, compared)
	assert.Positive(t, reversed)
}

func TestADotOnOneSideMakesTheOlderVersion(t *testing.T) {
	// act
	compared := vercmp.Compare("1a.0", "1a0")

	// assert
	assert.Negative(t, compared)
}

func TestATildeMakesTheOlderVersionEvenAgainstTheEnd(t *testing.T) {
	// act
	compared := vercmp.Compare("123~rc1", "123")
	reversed := vercmp.Compare("123", "123~rc1")

	// assert
	assert.Negative(t, compared)
	assert.Positive(t, reversed)
}

func TestLettersCompareInAlphabeticalOrder(t *testing.T) {
	// act
	compared := vercmp.Compare("123.a", "123.b")

	// assert
	assert.Negative(t, compared)
}

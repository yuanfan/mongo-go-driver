package aggregateopt

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/internal/testutil/helpers"
)

func createNestedBundle1(t *testing.T) *AggregateBundle {
	nestedBundle := BundleAggregate(AllowDiskUse(false), Comment("hello world nested"))
	testhelpers.RequireNotNil(t, nestedBundle, "nested bundle was nil")

	outerBundle := BundleAggregate(AllowDiskUse(true), MaxTime(500), nestedBundle, BatchSize(1000))
	testhelpers.RequireNotNil(t, outerBundle, "outer bundle was nil")

	return outerBundle
}

// Test doubly nested bundle
func createNestedBundle2(t *testing.T) *AggregateBundle {
	b1 := BundleAggregate(AllowDiskUse(false), Comment("nest1"))
	testhelpers.RequireNotNil(t, b1, "nested bundle was nil")

	b2 := BundleAggregate(MaxTime(100), b1, Comment("nest2"))
	testhelpers.RequireNotNil(t, b2, "nested bundle was nil")

	outerBundle := BundleAggregate(AllowDiskUse(true), MaxTime(500), b2, BatchSize(1000))
	testhelpers.RequireNotNil(t, outerBundle, "outer bundle was nil")

	return outerBundle
}

// Test two top level nested bundles
func createNestedBundle3(t *testing.T) *AggregateBundle {
	b1 := BundleAggregate(AllowDiskUse(false), Comment("nest1"))
	testhelpers.RequireNotNil(t, b1, "nested bundle was nil")

	b2 := BundleAggregate(MaxTime(100), b1, Comment("nest2"))
	testhelpers.RequireNotNil(t, b2, "nested bundle was nil")

	b3 := BundleAggregate(AllowDiskUse(true), Comment("nest3"))
	testhelpers.RequireNotNil(t, b3, "nested bundle was nil")

	b4 := BundleAggregate(MaxTime(100), b3, Comment("nest4"))
	testhelpers.RequireNotNil(t, b4, "nested bundle was nil")

	outerBundle := BundleAggregate(b4, MaxTime(500), b2, BatchSize(1000))
	testhelpers.RequireNotNil(t, outerBundle, "outer bundle was nil")

	return outerBundle
}

func TestAggregateOpt(t *testing.T) {
	var bundle1 *AggregateBundle
	bundle1 = bundle1.BypassDocumentValidation(true).Comment("hello world").BypassDocumentValidation(false)
	testhelpers.RequireNotNil(t, bundle1, "created bundle was nil")
	bundle1Opts := []option.Optioner{
		BypassDocumentValidation(true).ConvertOption(),
		Comment("hello world").ConvertOption(),
		BypassDocumentValidation(false).ConvertOption(),
	}
	bundle1DedupOpts := []option.Optioner{
		Comment("hello world").ConvertOption(),
		BypassDocumentValidation(false).ConvertOption(),
	}

	bundle2 := BundleAggregate(BatchSize(1))
	bundle2Opts := []option.Optioner{
		OptBatchSize(1).ConvertOption(),
	}

	bundle3 := BundleAggregate().
		BatchSize(1).
		Comment("Hello").
		BatchSize(2).
		BypassDocumentValidation(false).
		BypassDocumentValidation(true).
		Comment("World")

	bundle3Opts := []option.Optioner{
		OptBatchSize(1).ConvertOption(),
		OptComment("Hello").ConvertOption(),
		OptBatchSize(2).ConvertOption(),
		OptBypassDocumentValidation(false).ConvertOption(),
		OptBypassDocumentValidation(true).ConvertOption(),
		OptComment("World").ConvertOption(),
	}

	bundle3DedupOpts := []option.Optioner{
		OptBatchSize(2).ConvertOption(),
		OptBypassDocumentValidation(true).ConvertOption(),
		OptComment("World").ConvertOption(),
	}

	nilBundle := BundleAggregate()
	var nilBundleOpts []option.Optioner

	nestedBundle1 := createNestedBundle1(t)
	nestedBundleOpts1 := []option.Optioner{
		OptAllowDiskUse(true).ConvertOption(),
		OptMaxTime(500).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("hello world nested").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}
	nestedBundleDedupOpts1 := []option.Optioner{
		OptMaxTime(500).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("hello world nested").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}

	nestedBundle2 := createNestedBundle2(t)
	nestedBundleOpts2 := []option.Optioner{
		OptAllowDiskUse(true).ConvertOption(),
		OptMaxTime(500).ConvertOption(),
		OptMaxTime(100).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("nest1").ConvertOption(),
		OptComment("nest2").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}
	nestedBundleDedupOpts2 := []option.Optioner{
		OptMaxTime(100).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("nest2").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}

	nestedBundle3 := createNestedBundle3(t)
	nestedBundleOpts3 := []option.Optioner{
		OptMaxTime(100).ConvertOption(),
		OptAllowDiskUse(true).ConvertOption(),
		OptComment("nest3").ConvertOption(),
		OptComment("nest4").ConvertOption(),
		OptMaxTime(500).ConvertOption(),
		OptMaxTime(100).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("nest1").ConvertOption(),
		OptComment("nest2").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}
	nestedBundleDedupOpts3 := []option.Optioner{
		OptMaxTime(100).ConvertOption(),
		OptAllowDiskUse(false).ConvertOption(),
		OptComment("nest2").ConvertOption(),
		OptBatchSize(1000).ConvertOption(),
	}

	t.Run("MakeOptions", func(t *testing.T) {
		head := bundle1

		bundleLen := 0
		for head != nil && head.option != nil {
			bundleLen++
			head = head.next
		}

		if bundleLen != 3 {
			t.Errorf("expected bundle length 3. got: %d", bundleLen)
		}
	})

	t.Run("Unbundle", func(t *testing.T) {
		var cases = []struct {
			name         string
			dedup        bool
			bundle       *AggregateBundle
			expectedOpts []option.Optioner
		}{
			{"NilBundle", false, nilBundle, nilBundleOpts},
			{"Bundle1", false, bundle1, bundle1Opts},
			{"Bundle1Dedup", true, bundle1, bundle1DedupOpts},
			{"Bundle2", false, bundle2, bundle2Opts},
			{"Bundle2Dedup", true, bundle2, bundle2Opts},
			{"Bundle3", false, bundle3, bundle3Opts},
			{"Bundle3Dedup", true, bundle3, bundle3DedupOpts},
			{"NestedBundle1_DedupFalse", false, nestedBundle1, nestedBundleOpts1},
			{"NestedBundle1_DedupTrue", true, nestedBundle1, nestedBundleDedupOpts1},
			{"NestedBundle2_DedupFalse", false, nestedBundle2, nestedBundleOpts2},
			{"NestedBundle2_DedupTrue", true, nestedBundle2, nestedBundleDedupOpts2},
			{"NestedBundle3_DedupFalse", false, nestedBundle3, nestedBundleOpts3},
			{"NestedBundle3_DedupTrue", true, nestedBundle3, nestedBundleDedupOpts3},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				options, err := tc.bundle.Unbundle(tc.dedup)
				testhelpers.RequireNil(t, err, "got non-nill error from unbundle: %s", err)

				if len(options) != len(tc.expectedOpts) {
					t.Errorf("options length does not match expected length. got %d expected %d", len(options),
						len(tc.expectedOpts))
				} else {
					for i, opt := range options {
						if opt != tc.expectedOpts[i] {
							t.Errorf("expected: %s\nreceived: %s", opt, tc.expectedOpts[i])
						}
					}
				}
			})
		}
	})
}
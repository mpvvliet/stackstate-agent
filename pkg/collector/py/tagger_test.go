// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build cpython

// NOTICE: See TestMain function in `utils_test.go` for Python initialization
package py

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/errors"
	"github.com/StackVista/stackstate-agent/pkg/tagger"
	"github.com/StackVista/stackstate-agent/pkg/tagger/collectors"
)

type DummyCollector struct {
}

func (c *DummyCollector) Detect(out chan<- []*collectors.TagInfo) (collectors.CollectionMode, error) {
	return collectors.FetchOnlyCollection, nil
}
func (c *DummyCollector) Fetch(entity string) ([]string, []string, error) {
	if entity == "404" {
		return nil, nil, errors.NewNotFound(entity)
	} else {
		return []string{entity + ":low"}, []string{entity + ":high", "other_tag:high"}, nil
	}
}

func TestGetTags(t *testing.T) {
	collectors.DefaultCatalog = map[string]collectors.CollectorFactory{
		"dummy": func() collectors.Collector {
			return &DummyCollector{}
		},
	}
	tagger.Init()

	// Make sure tagger works as expected first
	low, err := tagger.Tag("test_entity", false)
	require.NoError(t, err)
	require.Equal(t, low, []string{"test_entity:low"})
	high, err := tagger.Tag("test_entity", true)
	require.NoError(t, err)
	assert.ElementsMatch(t, high, []string{"test_entity:low", "test_entity:high", "other_tag:high"})

	check, _ := getCheckInstance("testtagger", "TestCheck")
	mockSender := mocksender.NewMockSender(check.ID())
	mockSender.SetupAcceptAll()

	err = check.Run()
	require.NoError(t, err)

	mockSender.AssertMetricTaggedWith(t, "Gauge", "metric.low_card", []string{"test_entity:low"})
	mockSender.AssertMetricTaggedWith(t, "Gauge", "metric.high_card", []string{"test_entity:low", "test_entity:high", "other_tag:high"})
	mockSender.AssertMetricTaggedWith(t, "Gauge", "metric.unknown", []string{})
}

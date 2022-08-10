package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateNextStep(t *testing.T) {
	t.Run("should return the last step if no interceptors are set", func(t *testing.T) {
		f := New(t, &SampleMigrator{})

		require.Equal(t, 7, f.calculateNextStep(7))
		f.currentStep = 5
		require.Equal(t, 7, f.calculateNextStep(7))
	})

	t.Run("should return the next step with an interceptor", func(t *testing.T) {
		interceptors := map[int]Interceptor{
			3:  func() error { return nil },
			12: func() error { return nil },
		}

		f := New(t, &SampleMigrator{}).
			RegisterInterceptors(interceptors)

		require.Equal(t, 3, f.calculateNextStep(7))
		f.currentStep = 5
		require.Equal(t, 7, f.calculateNextStep(7))
		require.Equal(t, 12, f.calculateNextStep(15))
		f.currentStep = 12
		require.Equal(t, 15, f.calculateNextStep(15))
	})
}

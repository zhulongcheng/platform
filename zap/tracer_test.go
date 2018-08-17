package zap

import (
	"testing"

	"github.com/influxdata/platform/snowflake"
	"go.uber.org/zap"
)

func BenchmarkStartSpan(b *testing.B) {
	tracer := &Tracer{
		Logger:      zap.NewNop(),
		IDGenerator: snowflake.NewIDGenerator(),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tracer.StartSpan("test")
	}
}

package parser

import (
	"github.com/turbonomic/prometurbo/prometurbo/pkg/discovery/exporter"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// Convert json metric response to EntityDTO
type EntityParser interface {
	parseMetric(metrics []*exporter.EntityMetric) ([]*proto.EntityDTO, error)
}

type GenericEntityParser struct {

}

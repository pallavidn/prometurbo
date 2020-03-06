package parser

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)


func convertJsonTypeToEntityType(jsonType string) proto.EntityDTO_EntityType {

	switch jsonType {
	case "APPLICATION":
		return proto.EntityDTO_BUSINESS_APPLICATION
	}

}

func convertJsonMetricTypeToCommodityType(jsonType string) proto.CommodityDTO_CommodityType {

	switch jsonType {
	case "CPU":
		return proto.CommodityDTO_CPU
	}

}
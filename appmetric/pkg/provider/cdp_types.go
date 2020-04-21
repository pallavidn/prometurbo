package provider

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
	"time"
)

// USING the common DIF Data

var DIFEntityType = map[proto.EntityDTO_EntityType]string{
	proto.EntityDTO_VIRTUAL_MACHINE:       "virtualMachine",
	proto.EntityDTO_APPLICATION_COMPONENT: "application",
	proto.EntityDTO_BUSINESS_APPLICATION:  "businessApplication",
	proto.EntityDTO_BUSINESS_TRANSACTION:  "businessTransaction",
	proto.EntityDTO_DATABASE_SERVER:       "databaseServer",
	proto.EntityDTO_SERVICE:               "service",
}

type CDPHostType string

const (
	VM        CDPHostType = "virtualMachine"
	CONTAINER CDPHostType = "container"
)

var CDPMetricType = map[proto.CommodityDTO_CommodityType]string{
	proto.CommodityDTO_RESPONSE_TIME:     "responseTime",
	proto.CommodityDTO_TRANSACTION:       "transaction",
	proto.CommodityDTO_VCPU:              "cpu",
	proto.CommodityDTO_VMEM:              "memory",
	proto.CommodityDTO_THREADS:           "threads",
	proto.CommodityDTO_HEAP:              "heap",
	proto.CommodityDTO_COLLECTION_TIME:   "collectionTime",
	proto.CommodityDTO_DB_MEM:            "dbMem",
	proto.CommodityDTO_DB_CACHE_HIT_RATE: "dbCacheHitRate",
	proto.CommodityDTO_CONNECTION: "connection",
}

// =============== DIF JSON Response from appMetric ========================

type DIFMetricResponse struct {
	Version    string `json:"version"`
	UpdateTime int64  `json:"updatetime"`
	Scope string `json:"scope"`
	Topology []*data.DIFEntity `json:"topology"`
}

func NewDIFMetricResponse() *DIFMetricResponse {
	return &DIFMetricResponse{
		Version:    "v1",
		UpdateTime: 0,
		Scope: "",
		Topology:   []*data.DIFEntity{},
	}
}

func (r *DIFMetricResponse) SetUpdateTime() {
	t := time.Now()
	r.UpdateTime = t.Unix()
}

func (r *DIFMetricResponse) SetScope(scope string) {
	r.Scope = scope
}

func (r *DIFMetricResponse) SetMetrics(dat []*data.DIFEntity) {
	r.Topology = dat
}

func (r *DIFMetricResponse) AddMetric(m *data.DIFEntity) {
	r.Topology = append(r.Topology, m)
}

// ======================== Convert EntityMetric to DIF Entities =============================

func ConvertToDIFMetric(m *EntityMetric) *data.DIFEntity {

	entityType, exists := DIFEntityType[m.Type]

	if !exists {
		glog.Errorf("Cannot find entity type for %v\n", m.Type)
	}
	cm := &data.DIFEntity{
		UID:    m.UID,
		Type:   entityType,
		Name:   m.UID,
		HostedOn:            nil,
		MatchingIdentifiers: nil,
		PartOf:              nil,
		Metrics:             nil,
	}


	//hostedOn
	hostedOn := &data.DIFHostedOn{}

	if m.HostedOnVM{
		fmt.Printf("Creating hosted on %s\n", m.UID)
		hostedOn.HostType = append(hostedOn.HostType, data.VM)
		hostedOn.IPAddress = m.UID
		cm.HostedOn = hostedOn
	}

	// partOf
	for key, label := range m.Labels {
		if key == "service" {
			parent := &data.DIFPartOf{
				ParentEntity: key,
				UniqueId:     label,
			}
			cm.PartOf = append(cm.PartOf, parent)
		}
		if key == "ip" {
			matchingIds := &data.DIFMatchingIdentifiers{
				IPAddress: m.UID,
			}
			cm.MatchingIdentifiers = matchingIds
		}
	}

	// metrics
	//var cdpMetrics []map[string][]*data.DIFMetricVal
	meMap := make(map[string][]*data.DIFMetricVal)
	for comm, metric := range m.Metrics {
		var meList []*data.DIFMetricVal
		metricType, exists := CDPMetricType[comm]
		if !exists {
			glog.Errorf("Cannot find metric type for comm %v\n", comm)
		}

		me := &data.DIFMetricVal{}

		for key, val := range metric {
			if key == "used" {
				me.Average = &val
			} else if key == "capacity" {
				me.Capacity = &val
			}
		}

		meList = append(meList, me)

		meMap[metricType] = meList
		//cdpMetrics = append(cdpMetrics, meMap)
	}

	cm.Metrics = meMap
	return cm
}

func CreateDIFServiceMetric(svcName string, metrics map[string]*EntityMetric) *data.DIFEntity {

	entityType := "service"

	cm := &data.DIFEntity{
		UID:  svcName,
		Type: entityType,
		Name: svcName,
	}

	ServicePrefix := "Service-"
	// stitching identifiers
	var matchingIds *data.DIFMatchingIdentifiers
	var svcIPs []string
	for svcIP, _ := range metrics {
		svcIPs = append(svcIPs, ServicePrefix+svcIP)
	}
	matchingIds = &data.DIFMatchingIdentifiers{
		IPAddress: strings.Join(svcIPs, ","),
	}
	cm.MatchingIdentifiers = matchingIds

	//var svcMetricsMap map[string][]*data.DIFMetricVal
	svcMetricsMap := make(map[string][]*data.DIFMetricVal)
	for _, m := range metrics {
		//cm.Source = m.Source
		for comm, metric := range m.Metrics {
			var meList []*data.DIFMetricVal
			metricType, exists := CDPMetricType[comm]
			if !exists {
				fmt.Printf("Cannot find metric type for comm %v\n", comm)
			}

			me := &data.DIFMetricVal{}

			for key, val := range metric {
				if key == "used" {
					me.Average = &val
				} else if key == "capacity" {
					me.Capacity = &val
				}
			}

			if _, exists := svcMetricsMap[metricType]; !exists {
				svcMetricsMap[metricType] = []*data.DIFMetricVal{}
			}
			meList = svcMetricsMap[metricType]
			meList = append(meList, me)
			svcMetricsMap[metricType] = meList
		}
	}

	//var cdpMetrics []map[string][]*data.DIFMetricVal
	//for metricType, meList := range svcMetricsMap {
	//	meMap := make(map[string][]*data.DIFMetricVal)
	//	meMap[metricType] = meList
	//	cdpMetrics = append(cdpMetrics, meMap)
	//}
	cm.Metrics = svcMetricsMap

	//fmt.Printf("%s --> %++v\n", cm.Source, cm)
	return cm
}

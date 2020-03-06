package provider

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"

	"time"
)

type CDPEntity struct {
	UID  string `json:"uniqueId"`
	Type string `json:"type"`
	Name string `json:"name"`

	HostedOn *CDPHostedOn	`json:"hostedOn"`
	MatchingIdentifiers *CDPMatchingIdentifiers `json:"matchIdentifiers"`
	PartOf []*CDPPartOf	`json:"partOf"`
	Metrics             []map[string][]*CDPMetricEntry	`json:"metrics,omitempty"`
	//Metrics  map[string][]*CDPMetricEntry `json:"metrics,omitempty"`
}

type CDPMatchingIdentifiers struct {
	IPAddress string 	`json:"ipAddress"`
}

type CDPHostedOn struct {
	HostType  []string `json:"hostType"`
	IPAddress string   `json:"ipAddress"`
	HostUuid string		`json:"hostUuid"`
}

type CDPPartOf struct {
	ParentEntity string	`json:"entity"`
	UniqueId string 	`json:"uniqueId"`
}

type CDPMetric struct {
	Metrics map[string]*CDPMetricEntry `json:"metrics,omitempty"`
}

type CDPMetricEntryBasic interface {
	GetAverage() float64
}

type CDPMetricEntryWithKey interface {
	CDPMetricEntryBasic
	GetKey() string
}

type CDPMetricWithRawData interface {

}

type CDPMetricEntry struct {
	Average  float64       `json:"average"`
	Min      float64       `json:"min"`
	Max      float64       `json:"min"`
	Capacity float64       `json:"min"`
	Unit     CDPMetricUnit `json:"unit"`
	Key      string        `json:"key"`
}

func (m *CDPMetricEntry) GetAverage() float64 {
	return m.Average
}

func (m *CDPMetricEntry) GetKey() string {
	return m.Key
}

type CDPMetricUnit string

const (
	COUNT CDPMetricUnit = "count"
	TPS   CDPMetricUnit = "tps"
	MS   CDPMetricUnit  = "ms"
	MB   CDPMetricUnit  = "mb"
	MHZ   CDPMetricUnit = "mhz"
	PCT   CDPMetricUnit = "pct"
)

var CDPEntityType = map[proto.EntityDTO_EntityType]string{
	proto.EntityDTO_VIRTUAL_MACHINE:          "virtualMachine",
	proto.EntityDTO_APPLICATION:          "application",
	proto.EntityDTO_BUSINESS_APPLICATION: "businessApplication",
	proto.EntityDTO_DATABASE_SERVER:      "databaseServer",
	proto.EntityDTO_VIRTUAL_APPLICATION:  "service",
}

type CDPHostType string
const (
	VM CDPHostType = "virtualMachine"
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
}

type CDPMetricResponse struct {
	Version    string `json:"version"`
	UpdateTime int64  `json:"updatetime"`

	Topology []*CDPEntity `json:"topology"`
}

func NewCDPMetricResponse() *CDPMetricResponse {
	return &CDPMetricResponse{
		Version:    "v1",
		UpdateTime: 0,
		Topology:   []*CDPEntity{},
	}
}

func (r *CDPMetricResponse) SetUpdateTime() {
	t := time.Now()
	r.UpdateTime = t.Unix()
}

func (r *CDPMetricResponse) SetMetrics(dat []*CDPEntity) {
	r.Topology = dat
}

func (r *CDPMetricResponse) AddMetric(m *CDPEntity) {
	r.Topology = append(r.Topology, m)
}

func ConvertToCDPMetric(m *EntityMetric) *CDPEntity {
	entityType, exists := CDPEntityType[m.Type]
	if !exists {
		glog.Errorf("Cannot find entity type for %v\n", m.Type)
	}
	cm := &CDPEntity{
		UID:  m.UID,
		Type: entityType,
		Name: m.UID,
	}
	// stitching identifiers
	matchingIds := &CDPMatchingIdentifiers{
		IPAddress:m.UID,
	}
	cm.MatchingIdentifiers = matchingIds

	//hostedOn
	hostedOn := &CDPHostedOn{}
	if m.HostedOnVM {
		fmt.Printf("Creating hosted on %s\n", m.UID)
		hostedOn.HostType = append(hostedOn.HostType, string(VM))
		hostedOn.IPAddress = m.UID
		cm.HostedOn = hostedOn
	}

	// partOf
	for key, label := range m.Labels {
		if key == "business_app" {
			parent := &CDPPartOf{
				ParentEntity: key,
				UniqueId:  label,
			}
			cm.PartOf = append(cm.PartOf, parent)
		}
	}
	// hosted on

	// metrics
	var cdpMetrics []map[string][]*CDPMetricEntry//make(map[string][]*CDPMetricEntry)
	for comm, metric := range m.Metrics {
		var meList []*CDPMetricEntry
		metricType, exists := CDPMetricType[comm]
		if !exists {
			glog.Errorf("Cannot find metric type for comm %v\n", comm)
		}

		me := &CDPMetricEntry{}

		for key, val := range metric {
			if key == "used" {
				me.Average = val
			} else if key == "capacity" {
				me.Capacity = val
			}
		}

		meList = append(meList, me)
		meMap := make(map[string][]*CDPMetricEntry)
		meMap[metricType] = meList
		cdpMetrics = append(cdpMetrics, meMap)
	}

	cm.Metrics = cdpMetrics
	return cm
}

func CreateCDPServiceMetric(m *EntityMetric) *CDPEntity {
	entityType, exists := CDPEntityType[m.Type]
	if !exists {
		glog.Errorf("Cannot find entity type for %v\n", m.Type)
	}
	if entityType != "application" {
		//glog.Errorf("Need application entity to create service entity %s\n", entityType)
		return nil
	}
	entityType = "service"
	cm := &CDPEntity{
		UID:  m.UID,
		Type: entityType,
		Name: m.UID,
	}
	// stitching identifiers
	matchingIds := &CDPMatchingIdentifiers{
		IPAddress:m.UID,
	}
	cm.MatchingIdentifiers = matchingIds

	// metrics
	var cdpMetrics []map[string][]*CDPMetricEntry
	for comm, metric := range m.Metrics {
		var meList []*CDPMetricEntry
		metricType, exists := CDPMetricType[comm]
		if !exists {
			fmt.Printf("Cannot find metric type for comm %v\n", comm)
		}

		me := &CDPMetricEntry{}

		for key, val := range metric {
			if key == "used" {
				me.Average = val
			} else if key == "capacity" {
				me.Capacity = val
			}
		}

		meList = append(meList, me)
		meMap := make(map[string][]*CDPMetricEntry)
		meMap[metricType] = meList
		cdpMetrics = append(cdpMetrics, meMap)
	}

	cm.Metrics = cdpMetrics
	return cm
}
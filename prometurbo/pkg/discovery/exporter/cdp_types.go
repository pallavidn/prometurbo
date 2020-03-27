package exporter

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

	HostedOn            *CDPHostedOn                   `json:"hostedOn"`
	MatchingIdentifiers *CDPMatchingIdentifiers        `json:"matchIdentifiers"`
	PartOf              []*CDPPartOf                   `json:"partOf"`
	Metrics             []map[string][]*CDPMetricEntry `json:"metrics,omitempty"`
	//Metrics  map[string][]*CDPMetricEntry `json:"metrics,omitempty"`
}

type CDPMatchingIdentifiers struct {
	IPAddress string `json:"ipAddress"`
}

type CDPHostedOn struct {
	HostType  []string `json:"hostType"`
	IPAddress string   `json:"ipAddress"`
	HostUuid  string   `json:"hostUuid"`
}

type CDPPartOf struct {
	ParentEntity string `json:"entity"`
	UniqueId     string `json:"uniqueId"`
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
	MS    CDPMetricUnit = "ms"
	MB    CDPMetricUnit = "mb"
	MHZ   CDPMetricUnit = "mhz"
	PCT   CDPMetricUnit = "pct"
)

var CDPEntityType = map[proto.EntityDTO_EntityType]string{
	proto.EntityDTO_VIRTUAL_MACHINE:      "virtualMachine",
	proto.EntityDTO_APPLICATION_COMPONENT:          "application",
	proto.EntityDTO_BUSINESS_APPLICATION: "businessApplication",
	proto.EntityDTO_DATABASE_SERVER:      "databaseServer",
	proto.EntityDTO_SERVICE:  "service",
}

var ReverseCDPEntityType = map[string]proto.EntityDTO_EntityType{
	"virtualMachine":      proto.EntityDTO_VIRTUAL_MACHINE,
	"application":         proto.EntityDTO_APPLICATION_COMPONENT,
	"businessApplication": proto.EntityDTO_BUSINESS_APPLICATION,
	"businessTransaction": proto.EntityDTO_BUSINESS_TRANSACTION,
	"databaseServer":      proto.EntityDTO_DATABASE_SERVER,
	"service":             proto.EntityDTO_SERVICE,
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
}

var ReverseCDPMetricType = map[string]proto.CommodityDTO_CommodityType{
	"responseTime":   proto.CommodityDTO_RESPONSE_TIME,
	"transaction":    proto.CommodityDTO_TRANSACTION,
	"cpu":            proto.CommodityDTO_VCPU,
	"memory":         proto.CommodityDTO_VMEM,
	"threads":        proto.CommodityDTO_THREADS,
	"heap":           proto.CommodityDTO_HEAP,
	"collectionTime": proto.CommodityDTO_COLLECTION_TIME,
	"dbMem":          proto.CommodityDTO_DB_MEM,
	"dbCacheHitRate": proto.CommodityDTO_DB_CACHE_HIT_RATE,
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

func ConvertFromCDPMetric(m *CDPEntity) *EntityMetric {
	if m.Type == "service" || m.Type == "virtualMachine" {
		glog.Errorf("Skip vm/service entity %s\n", m.Type)
		return nil
	}

	entityType, exists := ReverseCDPEntityType[m.Type]
	if !exists {
		glog.Errorf("Cannot find entity type for %v\n", m.Type)
	}

	em := &EntityMetric{
		UID:        m.UID,
		Type:       entityType,
		Labels:     nil,
		Metrics:    nil,
		HostedOnVM: false,
	}

	if m.HostedOn != nil {
		em.HostedOnVM = true
	}

	if m.PartOf != nil {
		partOf := m.PartOf[0]
		em.Labels = make(map[string]string)
		em.Labels[partOf.ParentEntity] = partOf.UniqueId
	}

	allComms := make(map[proto.CommodityDTO_CommodityType]map[string]float64)
	for _, metricMap := range m.Metrics {
		for metricName, metricList := range metricMap {
			commType, exists := ReverseCDPMetricType[metricName]
			if !exists {
				glog.Errorf("Cannot find metric type for comm %v\n", metricName)
			}
			//Metrics    map[proto.CommodityDTO_CommodityType]map[string]float64
			commValues := make(map[string]float64)
			for _, metric := range metricList {
				commValues["capacity"] = metric.Capacity
				commValues["used"] = metric.Average
			}
			allComms[commType] = commValues
		}
	}
	em.Metrics = allComms

	return em
}

// ============================================================
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
		IPAddress: m.UID,
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
				UniqueId:     label,
			}
			cm.PartOf = append(cm.PartOf, parent)
		}
	}

	// metrics
	var cdpMetrics []map[string][]*CDPMetricEntry //make(map[string][]*CDPMetricEntry)
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
		glog.Errorf("Need application entity to create service entity %s\n", entityType)
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
		IPAddress: m.UID,
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

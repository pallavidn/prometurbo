package provider

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/prometurbo/appmetric/pkg/config"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type BusinessTopologyEditor struct {
	BizAppConfBySource config.BusinessAppConfBySource
}

func (b *BusinessTopologyEditor) BuildCDPEntities() []*data.DIFEntity {
	var bizEntities []*data.DIFEntity

	transToAppsMap := make(map[string][]string)
	svcToTransMap := make(map[string][]string)
	svcToAppsMap := make(map[string][]string)

	for source, bizAppConfByName := range b.BizAppConfBySource {
		for name, bizAppConf := range bizAppConfByName {
			glog.Infof("Source %s Name %s BizApp %v", source, name, bizAppConf)
			bizAppId := fmt.Sprintf("%s:%s", bizAppConf.Name, bizAppConf.From)

			for _, service := range bizAppConf.Services {
				if _, exists := svcToAppsMap[service]; exists {
					svcToAppsMap[service] = []string{}
				}

				svcToAppsMap[service] = append(svcToAppsMap[service], bizAppId)
			}
			for _, trans := range bizAppConf.Transactions {
				if _, exists := transToAppsMap[trans.Path]; exists {
					transToAppsMap[trans.Path] = []string{}
				}
				transToAppsMap[trans.Path] = append(transToAppsMap[trans.Path], bizAppId)
			}

			for _, bizTrans := range bizAppConf.Transactions {
				for _, service := range bizTrans.DependOn {
					if _, exists := svcToTransMap[service]; exists {
						svcToTransMap[service] = []string{}
					}
					svcToTransMap[service] = append(svcToTransMap[service], bizTrans.Path)
				}
			}

			bizAppEntity := BizAppToCDPMetric(bizAppConf)
			bizEntities = append(bizEntities, bizAppEntity)
		}
	}

	for _, bizAppConfByName := range b.BizAppConfBySource {
		for _, bizAppConf := range bizAppConfByName {
			for _, bizTrans := range bizAppConf.Transactions {
				bizTransEntity := BizTransToCDPMetric(bizTrans, transToAppsMap[bizTrans.Path])
				if bizTransEntity == nil {
					glog.Infof("### NIL BIZ TRANS for %s\n", bizTrans.Path)
				} else {
					glog.Infof("### %s", bizTransEntity.String())
				}
				bizEntities = append(bizEntities, bizTransEntity)
			}
		}
	}

	for svcName, bizApps := range svcToAppsMap {
		bizTxs := svcToTransMap[svcName]
		svcEntity := ServiceToCDPMetric(svcName, bizApps, bizTxs)
		bizEntities = append(bizEntities, svcEntity)
	}

	return bizEntities
}

func BizAppToCDPMetric(bizApp *config.BusinessApplication) *data.DIFEntity {
	cm := &data.DIFEntity{
		UID:  fmt.Sprintf("%s:%s", bizApp.Name, bizApp.From),
		Type: "businessApplication",
		Name: bizApp.Name,
	}

	return cm
}

func ServiceToCDPMetric(service string, bizApps, bizTxs []string) *data.DIFEntity {
	cm := &data.DIFEntity{
		UID:  service,
		Type: "service",
		Name: service,
	}
	for _, bizApp := range bizApps {
		parent := &data.DIFPartOf{
			ParentEntity: "businessApplication",
			UniqueId:     bizApp,
		}
		cm.PartOf = append(cm.PartOf, parent)
	}
	for _, bizTx := range bizTxs {
		parent := &data.DIFPartOf{
			ParentEntity: "businessTransaction",
			UniqueId:     bizTx,
		}
		cm.PartOf = append(cm.PartOf, parent)
	}

	return cm
}

func BizTransToCDPMetric(bizTrans config.Transaction, bizApps []string) *data.DIFEntity {
	var transEntity *data.DIFEntity
	name := bizTrans.Name
	if name == "" {
		name = bizTrans.Path
	}
	transEntity = &data.DIFEntity{
		UID:  bizTrans.Path,
		Type: "businessTransaction",
		Name: name,
	}
	if bizApps != nil {
		for _, bizApp := range bizApps {
			parent := &data.DIFPartOf{
				ParentEntity: "businessApplication",
				UniqueId:     bizApp,
			}
			transEntity.PartOf = append(transEntity.PartOf, parent)
		}
	}

	return transEntity
}
package rdb

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/scaleway/scaleway-operator/pkg/manager/scaleway"
	"github.com/scaleway/scaleway-operator/pkg/utils"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdbv1alpha1 "github.com/scaleway/scaleway-operator/apis/rdb/v1alpha1"
)

// InstanceManager manages the RDB instances
type InstanceManager struct {
	client.Client
	API *rdb.API
	scaleway.Manager
	Log logr.Logger
}

// Ensure reconciles the RDB instance resource
func (m *InstanceManager) Ensure(ctx context.Context, obj runtime.Object) (bool, error) {
	instance, err := convertInstance(obj)
	if err != nil {
		return false, err
	}

	region := scw.Region(instance.Spec.Region)

	// if instanceID is empty, we need to create the instance
	if instance.Spec.InstanceID == "" {
		return false, m.createInstance(ctx, instance)
	}

	rdbInstanceResp, err := m.API.GetInstance(&rdb.GetInstanceRequest{
		Region:     region,
		InstanceID: instance.Spec.InstanceID,
	})
	if err != nil {
		return false, err
	}

	needReturn, err := m.updateInstance(instance, rdbInstanceResp)
	if err != nil {
		return false, err
	}
	if needReturn {
		return false, nil
	}

	needReturn, err = m.upgradeInstance(instance, rdbInstanceResp)
	if err != nil {
		return false, err
	}
	if needReturn {
		return false, nil
	}

	if instance.Spec.ACL != nil {
		err = m.updateACLs(ctx, instance, rdbInstanceResp)
		if err != nil {
			return false, err
		}
	}

	if rdbInstanceResp.Endpoint != nil {
		instance.Status.Endpoint.IP = rdbInstanceResp.Endpoint.IP.String()
		instance.Status.Endpoint.Port = int32(rdbInstanceResp.Endpoint.Port)
	}

	return rdbInstanceResp.Status == rdb.InstanceStatusReady, nil
}

// Delete deletes the RDB instance resource
func (m *InstanceManager) Delete(ctx context.Context, obj runtime.Object) (bool, error) {
	instance, err := convertInstance(obj)
	if err != nil {
		return false, err
	}

	region := scw.Region(instance.Spec.Region)

	resourceID := instance.Spec.InstanceID
	if resourceID == "" {
		return true, nil
	}

	_, err = m.API.DeleteInstance(&rdb.DeleteInstanceRequest{
		Region:     region,
		InstanceID: resourceID,
	})
	if err != nil {
		if _, ok := err.(*scw.ResourceNotFoundError); ok {
			return true, nil
		}
		return false, err
	}

	//instance.Status.Status = strcase.ToCamel(instanceResp.Status.String())

	return false, nil
}

// GetOwners returns the owners of the RDB instance resource
func (m *InstanceManager) GetOwners(ctx context.Context, obj runtime.Object) ([]scaleway.Owner, error) {
	return nil, nil
}

func (m *InstanceManager) createInstance(ctx context.Context, instance *rdbv1alpha1.RDBInstance) error {
	region := scw.Region(instance.Spec.Region)

	if instance.Spec.InstanceFrom != nil {
		instanceID, region, err := m.getInstanceFromIDAndRegion(ctx, instance)
		if err != nil {
			return err
		}
		rdbInstanceResp, err := m.API.CloneInstance(&rdb.CloneInstanceRequest{
			Region:     region,
			InstanceID: instanceID,
			NodeType:   scw.StringPtr(instance.Spec.NodeType),
			Name:       instance.Name,
		})
		if err != nil {
			return err
		}
		instance.Spec.InstanceID = rdbInstanceResp.ID
		instance.Spec.Region = rdbInstanceResp.Region.String()
		err = m.Client.Update(ctx, instance)
		if err != nil {
			return err
		}

		return nil
	}

	disableBackup := true
	if instance.Spec.AutoBackup != nil {
		disableBackup = instance.Spec.AutoBackup.Disabled
	}

	createRequest := &rdb.CreateInstanceRequest{
		Region:        region,
		DisableBackup: disableBackup,
		Engine:        instance.Spec.Engine,
		IsHaCluster:   instance.Spec.IsHaCluster,
		Name:          instance.Name,
		NodeType:      instance.Spec.NodeType,
		Tags:          utils.LabelsToTags(instance.Labels),
	}

	rdbInstanceResp, err := m.API.CreateInstance(createRequest)
	if err != nil {
		return err
	}
	instance.Spec.InstanceID = rdbInstanceResp.ID
	instance.Spec.Region = rdbInstanceResp.Region.String()
	err = m.Client.Update(ctx, instance)
	if err != nil {
		return err
	}

	return nil
}

func (m *InstanceManager) updateInstance(instance *rdbv1alpha1.RDBInstance, rdbInstance *rdb.Instance) (bool, error) {
	needsUpdate := false
	updateRequest := &rdb.UpdateInstanceRequest{
		Region:     scw.Region(instance.Spec.Region),
		InstanceID: instance.Spec.InstanceID,
	}

	if !utils.CompareTagsLabels(rdbInstance.Tags, instance.Labels) {
		updateRequest.Tags = scw.StringsPtr(utils.LabelsToTags(instance.Labels))
		needsUpdate = true
	}

	if instance.Spec.AutoBackup != nil && rdbInstance.BackupSchedule != nil {
		if instance.Spec.AutoBackup.Disabled != rdbInstance.BackupSchedule.Disabled {
			updateRequest.IsBackupScheduleDisabled = scw.BoolPtr(instance.Spec.AutoBackup.Disabled)
			needsUpdate = true
		}
		if instance.Spec.AutoBackup.Frequency != nil && uint32(*instance.Spec.AutoBackup.Frequency) != rdbInstance.BackupSchedule.Frequency {
			updateRequest.BackupScheduleFrequency = scw.Uint32Ptr(uint32(*instance.Spec.AutoBackup.Frequency))
			needsUpdate = true
		}
		if instance.Spec.AutoBackup.Retention != nil && uint32(*instance.Spec.AutoBackup.Retention) != rdbInstance.BackupSchedule.Retention {
			updateRequest.BackupScheduleRetention = scw.Uint32Ptr(uint32(*instance.Spec.AutoBackup.Retention))
			needsUpdate = true
		}
	}

	if needsUpdate {
		_, err := m.API.UpdateInstance(updateRequest)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (m *InstanceManager) upgradeInstance(instance *rdbv1alpha1.RDBInstance, rdbInstance *rdb.Instance) (bool, error) {
	upgradeRequest := &rdb.UpgradeInstanceRequest{
		Region:     scw.Region(instance.Spec.Region),
		InstanceID: instance.Spec.InstanceID,
	}

	if rdbInstance.IsHaCluster != instance.Spec.IsHaCluster {
		upgradeRequest.EnableHa = scw.BoolPtr(instance.Spec.IsHaCluster)
		_, err := m.API.UpgradeInstance(upgradeRequest)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	if rdbInstance.NodeType != instance.Spec.NodeType {
		upgradeRequest.NodeType = scw.StringPtr(instance.Spec.NodeType)
		_, err := m.API.UpgradeInstance(upgradeRequest)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (m *InstanceManager) getNodesIP(ctx context.Context) ([]net.IPNet, error) {
	var nodesIP []net.IPNet

	nodesList := corev1.NodeList{}
	err := m.Client.List(ctx, &nodesList)
	if err != nil {
		return nil, err
	}
	for _, node := range nodesList.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeExternalIP || addr.Type == corev1.NodeInternalIP {
				parsedIP := net.ParseIP(addr.Address)
				if parsedIP != nil {
					nodesIP = append(nodesIP, getIPNetFromIP(parsedIP))
				}
			}
		}
	}

	return nodesIP, nil
}

func checkRulesUpdate(instance *rdbv1alpha1.RDBInstance, existingACLs *rdb.ListInstanceACLRulesResponse, nodesIP []net.IPNet) bool {
	needRulesUpdate := len(existingACLs.Rules) != len(nodesIP)+len(instance.Spec.ACL.Rules)

	if !needRulesUpdate {
		for _, existingRule := range existingACLs.Rules {
			foundRule := false
			for _, wantedRule := range instance.Spec.ACL.Rules {
				wantedRuleParsed := net.ParseIP(wantedRule.IPRange)
				if wantedRuleParsed != nil && existingRule.IP.String() == wantedRuleParsed.String() {
					foundRule = true
					break
				}
			}
			for _, nodeRule := range nodesIP {
				if foundRule {
					break
				}

				if nodeRule.String() == existingRule.IP.String() {
					foundRule = true
				}
			}
			if !foundRule {
				needRulesUpdate = true
			}
		}
	}

	return needRulesUpdate
}

func (m *InstanceManager) updateACLs(ctx context.Context, instance *rdbv1alpha1.RDBInstance, rdbInstance *rdb.Instance) error {
	existingACLs, err := m.API.ListInstanceACLRules(&rdb.ListInstanceACLRulesRequest{
		Region:     scw.Region(instance.Spec.Region),
		InstanceID: instance.Spec.InstanceID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}

	var nodesIP []net.IPNet

	if instance.Spec.ACL.AllowCluster {
		nodesIP, err = m.getNodesIP(ctx)
		if err != nil {
			return err
		}
	}

	if checkRulesUpdate(instance, existingACLs, nodesIP) {
		rules := []*rdb.ACLRuleRequest{}
		for _, wantedRule := range instance.Spec.ACL.Rules {
			_, wantedRuleParsed, err := net.ParseCIDR(wantedRule.IPRange)
			if err != nil {
				m.Log.Error(err, "error parsing ip range, ignoring")
				continue
			}
			if wantedRuleParsed != nil {
				rules = append(rules, &rdb.ACLRuleRequest{
					IP: scw.IPNet{
						IPNet: *wantedRuleParsed,
					},
					Description: wantedRule.Description,
				})
			}
		}
		for _, nodeIP := range nodesIP {
			rules = append(rules, &rdb.ACLRuleRequest{
				IP: scw.IPNet{
					IPNet: nodeIP,
				},
				Description: "Kuberentes node",
			})
		}
		_, err = m.API.SetInstanceACLRules(&rdb.SetInstanceACLRulesRequest{
			Region:     scw.Region(instance.Spec.Region),
			InstanceID: instance.Spec.InstanceID,
			Rules:      rules,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func convertInstance(obj runtime.Object) (*rdbv1alpha1.RDBInstance, error) {
	instance, ok := obj.(*rdbv1alpha1.RDBInstance)
	if !ok {
		return nil, fmt.Errorf("failed type assertion on kind: %s", obj.GetObjectKind().GroupVersionKind().String())
	}
	return instance, nil
}

func getIPNetFromIP(ip net.IP) net.IPNet {
	ipNet := net.IPNet{
		IP: ip,
	}
	if ip.To4() != nil {
		ipNet.Mask = net.CIDRMask(32, 32)
	} else {
		ipNet.Mask = net.CIDRMask(128, 128)
	}
	return ipNet
}

func (m *InstanceManager) getInstanceFromIDAndRegion(ctx context.Context, instance *rdbv1alpha1.RDBInstance) (string, scw.Region, error) {
	instanceFrom := &rdbv1alpha1.RDBInstance{}

	if instance.Spec.InstanceFrom.Name != "" {
		instanceNamespace := instance.Spec.InstanceFrom.Namespace
		if instanceNamespace == "" {
			instanceNamespace = instance.Namespace
		}

		err := m.Get(ctx, client.ObjectKey{Name: instance.Spec.InstanceFrom.Name, Namespace: instanceNamespace}, instanceFrom)
		if err != nil {
			return "", "", err
		}

		return instanceFrom.Spec.InstanceID, scw.Region(instanceFrom.Spec.Region), nil
	}

	return instance.Spec.InstanceFrom.ExternalID, scw.Region(instance.Spec.InstanceFrom.Region), nil
}

package ibm

import (
	"github.com/IBM-Bluemix/bluemix-go/api/cf/cfv2"
	"github.com/IBM-Bluemix/bluemix-go/api/iampap/iampapv1"
	"github.com/hashicorp/terraform/helper/schema"

	"strings"
)

//HashInt ...
func HashInt(v interface{}) int { return v.(int) }

func expandStringList(input []interface{}) []string {
	vs := make([]string, len(input))
	for i, v := range input {
		vs[i] = v.(string)
	}
	return vs
}

func flattenStringList(list []string) []interface{} {
	vs := make([]interface{}, len(list))
	for i, v := range list {
		vs[i] = v
	}
	return vs
}

func expandIntList(input []interface{}) []int {
	vs := make([]int, len(input))
	for i, v := range input {
		vs[i] = v.(int)
	}
	return vs
}

func flattenIntList(list []int) []interface{} {
	vs := make([]interface{}, len(list))
	for i, v := range list {
		vs[i] = v
	}
	return vs
}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		out[i] = v
	}
	return schema.NewSet(f, out)
}

func flattenRoute(in []cfv2.Route) *schema.Set {
	vs := make([]string, len(in))
	for i, v := range in {
		vs[i] = v.GUID
	}
	return newStringSet(schema.HashString, vs)
}

func stringSliceToSet(in []string) *schema.Set {
	vs := make([]string, len(in))
	for i, v := range in {
		vs[i] = v
	}
	return newStringSet(schema.HashString, vs)
}

func flattenServiceBindings(in []cfv2.ServiceBinding) *schema.Set {
	vs := make([]string, len(in))
	for i, v := range in {
		vs[i] = v.ServiceInstanceGUID
	}
	return newStringSet(schema.HashString, vs)
}

func flattenPort(in []int) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return schema.NewSet(HashInt, out)
}

func flattenIAMPolicyResource(list []iampapv1.Resources, iamClient iampapv1.IAMPAPAPI) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		serviceName, _ := iamClient.IAMService().GetServiceDispalyName(i.ServiceName)
		region := i.Region
		resourceType := i.ResourceType
		resource := i.Resource
		spaceId := i.SpaceId
		organizationId := i.OrganizationId
		serviceInstance := i.ServiceInstance
		ok, element := contains(serviceName, region, resourceType, resource, spaceId, organizationId, result)
		if !ok {
			l := map[string]interface{}{
				"service_name":      serviceName,
				"region":            region,
				"resource_type":     resourceType,
				"resource":          resource,
				"space_guid":        spaceId,
				"organization_guid": organizationId,
			}
			l["service_instance"] = []string{i.ServiceInstance}
			result = append(result, l)
		} else {
			v := element["service_instance"].([]string)
			v = append(v, serviceInstance)
			element["service_instance"] = v
		}
	}
	return result
}

func contains(name, region, resourceType, resource, spaceGuid, organizationGuid string, value []map[string]interface{}) (bool, map[string]interface{}) {
	for i := 0; i < len(value); i++ {
		m := value[i]
		if strings.Compare(m["service_name"].(string), name) == 0 && strings.Compare(m["region"].(string), region) == 0 && strings.Compare(m["resource_type"].(string), resourceType) == 0 && strings.Compare(m["resource"].(string), resource) == 0 && strings.Compare(m["space_guid"].(string), spaceGuid) == 0 && strings.Compare(m["organization_guid"].(string), organizationGuid) == 0 {
			return true, m
		}
	}
	return false, nil
}

func flattenIAMPolicyRoles(list []iampapv1.Roles, rolesMap map[string]string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"id": rolesMap[i.ID],
		}

		result = append(result, l)
	}
	return result
}

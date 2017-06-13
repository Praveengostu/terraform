package ibm

import (
	v1 "github.com/IBM-Bluemix/bluemix-go/api/iampap/iampapv1"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceIBMIAMAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMIAMAccessPolicyCreate,
		Read:     resourceIBMIAMAccessPolicyRead,
		Update:   resourceIBMIAMAccessPolicyUpdate,
		Delete:   resourceIBMIAMAccessPolicyDelete,
		Exists:   resourceIBMIAMAccessPolicyExists,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"account_guid": {
				Description: "The bluemix account guid",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"user_id": {
				Description: "id is the IBM id unique identifier (IUI). This id is the value in iam_id field from the users' IAM token.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_instance": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"space_guid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"organization_guid": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceIBMIAMAccessPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}
	accountGuid := d.Get("account_guid").(string)
	userId := d.Get("user_id").(string)
	roleIdSet := d.Get("roles").(*schema.Set)
	roleId := getRoleIDS(roleIdSet)
	roles := v1.Roles{
		ID: roleId,
	}

	resources := v1.Resources{
		AccountId: accountGuid,
	}
	if serviceName, ok := d.GetOk("service_name"); ok {
		resources.ServiceName = serviceName.(string)
	}
	if serviceInstance, ok := d.GetOk("service_instance"); ok {
		resources.ServiceInstance = serviceInstance.(string)
	}
	if region, ok := d.GetOk("region"); ok {
		resources.Region = region.(string)
	}
	if resourceType, ok := d.GetOk("resource_type"); ok {
		resources.ResourceType = resourceType.(string)
	}
	if resource, ok := d.GetOk("resource"); ok {
		resources.Resource = resource.(string)
	}
	if spaceGuid, ok := d.GetOk("space_guid"); ok {
		resources.SpaceId = spaceGuid.(string)
	}
	if organizationGuid, ok := d.GetOk("organization_guid"); ok {
		resources.OrganizationId = organizationGuid.(string)
	}
	params := v1.AccessPolicyRequest{
		Roles:     roles,
		Resources: resources,
	}

	accessPolicyResponse, etag, err := iamClient.IamPap().Create(accountGuid, userId, params)
	if err != nil {
		return err
	}
	d.SetId(accessPolicyResponse.ID)
	d.Set("roles", roleIdSet)
	d.Set("scope", accountGuid)
	d.Set("user_id", userId)
	d.Set("etag", etag)

	return resourceIBMIAMAccessPolicyRead(d, meta)
}

func getRoleIDS(roleIdSet *schema.Set) []string {
	roleIDS := make([]string, 0, roleIdSet.Len())
	for _, elem := range roleIdSet.List() {
		roleID := elem.(string)
		roleIDS = append(roleIDS, roleID)
	}
	return roleIDS
}

func resourceIBMIAMAccessPolicyRead(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}

	scope := d.Get("scope").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()

	iamPolicy, err := iamClient.IamPap().Find(scope, userId, policyId)
	if err != nil {
		return err
	}

	resources := iamPolicy.Resources
	d.Set("service_name", resources[0].ServiceName)
	d.Set("service_instance", resources[0].ServiceInstance)
	d.Set("region", resources[0].Region)
	d.Set("resource", resources[0].Resource)
	d.Set("resource_type", resources[0].ResourceType)
	d.Set("organization_guid", resources[0].OrganizationId)
	d.Set("space_guid", resources[0].SpaceId)
	d.Set("account_guid", resources[0].AccountId)

	return nil
}

func resourceIBMIAMAccessPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}

	scope := d.Get("scope").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()
	rolesSet := d.Get("roles").(*schema.Set)
	roles := getRoleIDS(rolesSet)
	etag := d.Get("etag").(string)
	/*if d.HasChange("roles") {
		o, n := d.GetChange("roles")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		add := ns.Difference(os)
		roles = getRoleIDS(add)
	}
	*/
	roleParam := v1.Roles{
		ID: roles,
	}

	resources := v1.Resources{}
	if d.HasChange("account_guid") {
		resources.AccountId = d.Get("account_guid").(string)
	}
	if d.HasChange("organization_guid") {
		resources.OrganizationId = d.Get("organization_guid").(string)
	}
	if d.HasChange("space_guid") {
		resources.SpaceId = d.Get("space_guid").(string)
	}
	if d.HasChange("resource") {
		resources.Resource = d.Get("resource").(string)
	}
	if d.HasChange("resource_type") {
		resources.ResourceType = d.Get("resource_type").(string)
	}
	if d.HasChange("region") {
		resources.Region = d.Get("region").(string)
	}
	if d.HasChange("service_instance") {
		resources.ServiceInstance = d.Get("service_instance").(string)
	}
	if d.HasChange("service_name") {
		resources.ServiceName = d.Get("service_name").(string)
	}

	accessPolicy := v1.AccessPolicyRequest{
		Roles:     roleParam,
		Resources: resources,
	}
	_, err = iamClient.IamPap().Update(scope, userId, policyId, etag, accessPolicy)
	d.Set("roles", rolesSet)
	return resourceIBMIAMAccessPolicyRead(d, meta)
}

func resourceIBMIAMAccessPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}

	scope := d.Get("scope").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()

	err = iamClient.IamPap().Delete(scope, userId, policyId)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceIBMIAMAccessPolicyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return false, err
	}

	scope := d.Get("scope").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()

	accessPolicyResponse, err := iamClient.IamPap().Find(scope, userId, policyId)
	if err != nil {
		return false, err
	}

	return policyId == accessPolicyResponse.ID, nil
}

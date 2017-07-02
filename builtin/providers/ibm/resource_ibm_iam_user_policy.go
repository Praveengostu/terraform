package ibm

import (
	"fmt"
	v1 "github.com/IBM-Bluemix/bluemix-go/api/iampap/iampapv1"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

const (
	VIEWER           = "viewer"
	EDITOR           = "editor"
	OPERATOR         = "getoperator"
	ADMINISTRATOR    = "administrator"
	VIEWER_ID        = "crn:v1:bluemix:public:iam::::role:Viewer"
	EDITOR_ID        = "crn:v1:bluemix:public:iam::::role:Editor"
	OPERATOR_ID      = "crn:v1:bluemix:public:iam::::role:Operator"
	ADMINISTRATOR_ID = "crn:v1:bluemix:public:iam::::role:Administrator"
)

const ALL_SERVICES = "All Identity and Access enbled services"

func resourceIBMIAMUserPolicy() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMIAMUserPolicyCreate,
		Read:     resourceIBMIAMUserPolicyRead,
		Update:   resourceIBMIAMUserPolicyUpdate,
		Delete:   resourceIBMIAMUserPolicyDelete,
		Exists:   resourceIBMIAMUserPolicyExists,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"account_guid": {
				Description: "The bluemix account guid",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"ibm_id": {
				Description: "The ibm id or email of user",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_instance": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
				},
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIBMIAMUserPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}
	accountGuid := d.Get("account_guid").(string)

	userEmail := d.Get("ibm_id").(string)
	userId, err := getIBMUniqueIdOfUser(accountGuid, userEmail, meta)
	if userId == "" || err != nil {
		return fmt.Errorf("User doesnot exist in the account", err)
	}

	roleIdSet := d.Get("roles").(*schema.Set)
	roles := getRoles(roleIdSet)

	policyServices := d.Get("resources").(*schema.Set)
	resources := createResources(policyServices, iamClient, accountGuid)

	params := v1.AccessPolicyRequest{
		Roles:     roles,
		Resources: resources,
	}

	accessPolicyResponse, etag, err := iamClient.IAMPolicy().Create(accountGuid, userId, params)
	if err != nil {
		return err
	}
	d.SetId(accessPolicyResponse.ID)
	d.Set("account_guid", accountGuid)
	d.Set("user_id", userId)
	d.Set("etag", etag)

	return resourceIBMIAMUserPolicyRead(d, meta)
}

func resourceIBMIAMUserPolicyRead(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}
	rolesMaps := make(map[string]string)
	rolesMaps[VIEWER_ID] = VIEWER
	rolesMaps[ADMINISTRATOR_ID] = ADMINISTRATOR
	rolesMaps[EDITOR_ID] = EDITOR
	rolesMaps[OPERATOR_ID] = OPERATOR

	accountGuid := d.Get("account_guid").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()
	iamPolicy, err := iamClient.IAMPolicy().Find(accountGuid, userId, policyId)
	if err != nil {
		return fmt.Errorf("Unable to read policy", err)
	}

	resources := iamPolicy.Resources
	roles := iamPolicy.Roles

	d.Set("roles", flattenIAMPolicyRoles(roles, rolesMaps))
	d.Set("resources", flattenIAMPolicyResource(resources, iamClient))

	return nil
}

func resourceIBMIAMUserPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}

	policyId := d.Id()
	accountGuid := d.Get("account_guid").(string)
	userId := d.Get("user_id").(string)
	etag := d.Get("etag").(string)

	var resources []v1.Resources
	var roles []v1.Roles
	if d.HasChange("roles") {
		_, newValue := d.GetChange("roles")
		newroles := newValue.(*schema.Set)
		roles = getRoles(newroles)
	}
	if d.HasChange("resources") {
		_, newValue := d.GetChange("resources")
		newResources := newValue.(*schema.Set)
		resources = createResources(newResources, iamClient, accountGuid)
	}

	if len(roles) > 0 && len(resources) == 0 {
		policyServices := d.Get("resources").(*schema.Set)
		resources = createResources(policyServices, iamClient, accountGuid)
	} else if len(roles) == 0 && len(resources) > 0 {
		roleIdSet := d.Get("roles").(*schema.Set)
		roles = getRoles(roleIdSet)
	}

	if len(roles) > 0 && len(resources) > 0 {
		accessPolicy := v1.AccessPolicyRequest{
			Roles:     roles,
			Resources: resources,
		}
		_, etag, err = iamClient.IAMPolicy().Update(accountGuid, userId, policyId, etag, accessPolicy)
		if err != nil {
			return fmt.Errorf("Unable to update policy1", err)
		}
		d.Set("account_guid", accountGuid)
		d.Set("etag", etag)
	}
	return resourceIBMIAMUserPolicyRead(d, meta)
}

func resourceIBMIAMUserPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return err
	}

	accountGuid := d.Get("account_guid").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()

	err = iamClient.IAMPolicy().Delete(accountGuid, userId, policyId)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceIBMIAMUserPolicyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	iamClient, err := meta.(ClientSession).IAMAPI()
	if err != nil {
		return false, err
	}

	accountGuid := d.Get("account_guid").(string)
	userId := d.Get("user_id").(string)
	policyId := d.Id()

	accessPolicyResponse, err := iamClient.IAMPolicy().Find(accountGuid, userId, policyId)
	if err != nil {
		return false, fmt.Errorf("Unable to check policy existence", err)
	}

	return policyId == accessPolicyResponse.ID, nil
}

func createResources(policyServices *schema.Set, iamClient v1.IAMPAPAPI, accountGuid string) []v1.Resources {
	var resources []v1.Resources
	for _, policyService := range policyServices.List() {
		rpm, _ := policyService.(map[string]interface{})
		serviceInstancesList := expandStringList(rpm["service_instance"].([]interface{}))
		serviceName, _ := iamClient.IAMService().GetServiceName(rpm["service_name"].(string))
		if len(serviceInstancesList) > 0 && strings.Compare(serviceName, ALL_SERVICES) != 0 {
			for _, serviceInstance := range serviceInstancesList {
				resources = append(resources, generateResource(rpm, serviceName, serviceInstance, accountGuid))
			}
		} else {
			resources = append(resources, generateResource(rpm, serviceName, "", accountGuid))
		}
	}
	return resources
}

func generateResource(rpm map[string]interface{}, serviceName, serviceInstance, accountGuid string) v1.Resources {
	resourceParam := v1.Resources{
		AccountId: accountGuid,
	}
	if strings.Compare(serviceName, ALL_SERVICES) != 0 {
		resourceParam.ServiceInstance = serviceInstance
		resourceParam.Region = rpm["region"].(string)
		resourceParam.ServiceName = serviceName
		resourceParam.ResourceType = rpm["resource_type"].(string)
		resourceParam.Resource = rpm["resource"].(string)
		resourceParam.SpaceId = rpm["space_guid"].(string)
		resourceParam.OrganizationId = rpm["organization_guid"].(string)
	}
	return resourceParam
}

func getIBMUniqueIdOfUser(accountGuid, userEmail string, meta interface{}) (string, error) {
	var ibmId string
	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return ibmId, err
	}
	accUsers, err := accountv1Client.Accounts().GetAccountUsers(accountGuid)
	if err != nil {
		return ibmId, err
	}
	for _, accUser := range accUsers {
		if strings.Compare(accUser.Email, userEmail) == 0 {
			return accUser.IbmUniqueId, nil

		}
	}

	return ibmId, nil
}

func getRoles(roleIdSet *schema.Set) []v1.Roles {
	rolesMaps := make(map[string]string)
	rolesMaps[VIEWER] = VIEWER_ID
	rolesMaps[ADMINISTRATOR] = ADMINISTRATOR_ID
	rolesMaps[EDITOR] = EDITOR_ID
	rolesMaps[OPERATOR] = OPERATOR_ID
	roleIDS := make([]v1.Roles, 0, roleIdSet.Len())
	for _, elem := range roleIdSet.List() {
		roleID := elem.(string)
		role := v1.Roles{
			ID: rolesMaps[strings.ToLower(roleID)],
		}
		roleIDS = append(roleIDS, role)
	}
	return roleIDS
}

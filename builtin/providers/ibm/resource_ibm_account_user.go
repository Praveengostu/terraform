package ibm

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceIBMAccountUser() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMAccountUserCreate,
		Read:     resourceIBMAccountUserRead,
		Delete:   resourceIBMAccountUserDelete,
		Exists:   resourceIBMAccountUserExists,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"account_guid": {
				Description: "The bluemix account guid",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"id": {
				Description: "id is the IBM id unique identifier (IUI). This id is the value in iam_id field from the users' IAM token.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user_email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org_guid": {
				Description: "The bluemix organization guid",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"org_role": {
				Description: "The bluemix organization role",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"space_guid": {
				Description: "The bluemix space guid",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"space_role": {
				Description: "The bluemix space role",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIBMAccountUserCreate(d *schema.ResourceData, meta interface{}) error {
	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return err
	}
	accountGuid := d.Get("account_guid").(string)
	userEmail := d.Get("user_email").(string)
	orgGuid := d.Get("org_guid").(string)
	spaceGuid := d.Get("space_guid").(string)
	orgRole := d.Get("org_role").(string)
	spaceRole := d.Get("space_role").(string)
	region := d.Get("region").(string)
	accInviteUserResp, err := accountv1Client.Accounts().InviteUser(accountGuid, userEmail, orgGuid, orgRole, spaceGuid, spaceRole, region)
	if err != nil {
		return err
	}
	d.SetId(accInviteUserResp.ID)
	d.Set("account_guid", accountGuid)

	return resourceIBMAccountUserRead(d, meta)
}

func resourceIBMAccountUserRead(d *schema.ResourceData, meta interface{}) error {
	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return err
	}
	accountGuid := d.Get("account_guid").(string)
	userId := d.Id()
	accUserResp, err := accountv1Client.Accounts().FindAccountUserByUserId(accountGuid, userId)
	d.Set("user_email", accUserResp.Email)
	d.Set("role", accUserResp.Role)
	d.Set("state", accUserResp.State)
	return nil
}

func resourceIBMAccountUserDelete(d *schema.ResourceData, meta interface{}) error {
	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return err
	}

	accountGuid := d.Get("account_guid").(string)
	userId := d.Id()

	err = accountv1Client.Accounts().DeleteAccountUser(accountGuid, userId)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceIBMAccountUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return false, err
	}
	accountGuid := d.Get("account_guid").(string)
	userId := d.Id()
	accUserResp, err := accountv1Client.Accounts().FindAccountUserByUserId(accountGuid, userId)
	if err != nil {
		return false, err
	}

	return accUserResp.Id == userId, nil
}

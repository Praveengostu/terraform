package ibm

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceIBMAccountUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMAccountUserRead,

		Schema: map[string]*schema.Schema{
			"org_guid": {
				Description: "The guid of the org",
				Type:        schema.TypeString,
				Required:    true,
			},
			"account_users": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ibm_uniqueid": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIBMAccountUserRead(d *schema.ResourceData, meta interface{}) error {
	accountv1Client, err := meta.(ClientSession).BluemixAcccountv1API()
	if err != nil {
		return err
	}

	accountGuid := d.Get("account_guid").(string)

	accountUsers, err := accountv1Client.Accounts().GetAccountUsers(accountGuid)
	if err != nil {
		return fmt.Errorf("Error retrieving users in account: %s", err)
	}
	accountUsersMap := make([]map[string]interface{}, 0)
	for _, user := range accountUsers {
		accountUser := make(map[string]interface{})
		accountUser["email"] = user.Email
		accountUser["state"] = user.State
		accountUser["role"] = user.Role
		accountUser["id"] = user.Id
		accountUser["ibm_uniqueid"] = user.IbmUniqueId
		accountUsersMap = append(accountUsersMap, accountUser)
	}
	d.Set("account_users", accountUsersMap)
	return nil
}

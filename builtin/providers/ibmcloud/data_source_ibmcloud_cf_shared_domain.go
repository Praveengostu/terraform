package ibmcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceIBMCloudCfSharedDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMCloudCfSharedDomainRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the shared domain",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateDomainName,
			},
		},
	}
}

func dataSourceIBMCloudCfSharedDomainRead(d *schema.ResourceData, meta interface{}) error {
	cfClient, err := meta.(ClientSession).CFAPI()
	if err != nil {
		return err
	}
	domainName := d.Get("name").(string)
	shdomain, err := cfClient.SharedDomains().FindByName(domainName)
	if err != nil {
		return fmt.Errorf("Error retrieving shared domain: %s", err)
	}
	d.SetId(shdomain.GUID)
	return nil

}

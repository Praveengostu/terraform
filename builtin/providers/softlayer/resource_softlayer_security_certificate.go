package softlayer

import (
	"fmt"
	"log"

	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

func resourceSoftLayerSecurityCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerSecurityCertificateCreate,
		Read:   resourceSoftLayerSecurityCertificateRead,
		Delete: resourceSoftLayerSecurityCertificateDelete,
		Exists: resourceSoftLayerSecurityCertificateExists,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},

			"certificate": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"private_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSoftLayerSecurityCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService

	if client == nil {
		return fmt.Errorf("The client is nil.")
	}

	template := datatypes.SoftLayer_Security_Certificate_Template{
		Certificate: d.Get("certificate").(string),
		PrivateKey:  d.Get("private_key").(string),
	}

	log.Printf("[INFO] Creating Security Certificate")

	cert, err := client.CreateSecurityCertificate(template)

	if err != nil {
		return fmt.Errorf("Error creating Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", cert.Id))

	return resourceSoftLayerSecurityCertificateRead(d, meta)
}

func resourceSoftLayerSecurityCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService
	if client == nil {
		return fmt.Errorf("The client is nil.")
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := client.GetObject(id)

	if err != nil {
		return fmt.Errorf("Unable to get Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", cert.Id))
	d.Set("certificate", cert.Certificate)
	d.Set("private_key", cert.PrivateKey)

	return nil
}

func resourceSoftLayerSecurityCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService

	if client == nil {
		return fmt.Errorf("The client was nil.")
	}

	_, err := client.DeleteObject(d.Get("id").(int))

	if err != nil {
		return fmt.Errorf("Error deleting Security Certificate %s: %s", d.Get("id"), err)
	}

	return nil
}

func resourceSoftLayerSecurityCertificateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client).securityCertificateService

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := client.GetObject(id)

	if err != nil {
		return false, fmt.Errorf("Error fetching Security Cerfiticate: %s", err)
	}

	return cert.Id == id && err == nil, nil
}

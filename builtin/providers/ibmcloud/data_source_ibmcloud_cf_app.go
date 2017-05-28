package ibmcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceIBMCloudCfApp() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIBMCloudCfAppRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name for the app",
			},
			"space_guid": {
				Description: "Define space guid to which app belongs",
				Type:        schema.TypeString,
				Required:    true,
			},
			"memory": {
				Description: "The amount of memory each instance should have. In megabytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"instances": {
				Description: "The number of instances",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"disk_quota": {
				Description: "The maximum amount of disk available to an instance of an app. In megabytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"command": {
				Description: "The initial command for the app",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"buildpack": {
				Description: "Buildpack to build the app. 3 options: a) Blank means autodetection; b) A Git Url pointing to a buildpack; c) Name of an installed buildpack.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"diego": {
				Description: "Use diego to stage and to run when available",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"environment_json": {
				Description: "Key/value pairs of all the environment variables to run in your app. Does not include any system or service variables.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			"ports": {
				Description: "Ports on which application may listen. Overwrites previously configured ports. Ports must be in range 1024-65535. Supported for Diego only.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Set: func(v interface{}) int {
					return v.(int)
				},
			},
			"route_guid": {
				Description: "Define the route guids which should be bound to the application.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"service_instance_guid": {
				Description: "Define the service instance guids that should be bound to this application.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"package_state": {
				Description: "The state of the application package whether staged, pending etc",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the application",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceIBMCloudCfAppRead(d *schema.ResourceData, meta interface{}) error {
	appClient := meta.(ClientSession).CloudFoundryAppClient()

	name := d.Get("name").(string)
	spaceGUID := d.Get("space_guid").(string)

	app, err := appClient.FindByName(spaceGUID, name)
	if err != nil {
		return err
	}

	d.SetId(app.GUID)
	d.Set("memory", app.Memory)
	d.Set("instances", app.Instances)
	d.Set("disk_quota", app.DiskQuota)
	d.Set("ports", flattenPort(app.Ports))
	if app.Command != nil {
		d.Set("command", app.Command)
	}

	if app.BuildPack != nil {
		d.Set("buildpack", app.BuildPack)
	}

	d.Set("diego", app.Diego)
	d.Set("environment_json", app.EnvironmentJSON)
	d.Set("package_state", app.PackageState)
	d.Set("state", app.State)
	d.Set("instances", app.Instances)

	route, err := appClient.ListRoutes(app.GUID)
	if err != nil {
		return err
	}
	if len(route) > 0 {
		d.Set("route_guid", flattenRoute(route))
	}
	svcBindings, err := appClient.ListServiceBindings(app.GUID)
	if err != nil {
		return err
	}
	if len(route) > 0 {
		d.Set("service_instance_guid", flattenServiceBindings(svcBindings))
	}
	return nil
}

package grafana

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	gapi "github.com/nytm/go-grafana-api"
)

func ResourceDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceDashboardCreate,
		Read:   resourceDashboardRead,
		Update: resourceDashboardUpdate,
		Delete: resourceDashboardDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"config_json": {
				Type:         schema.TypeString,
				Required:     true,
				StateFunc:    normalizeDashboardConfigJSON,
				ValidateFunc: validateDashboardConfigJSON,
			},
		},
	}
}

func resourceDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	dashboard := gapi.Dashboard{}

	dashboard.Model = prepareDashboardModel(d.Get("config_json").(string))

	dashboard.Folder = int64(d.Get("folder").(int))

	resp, err := client.NewDashboard(dashboard)
	if err != nil {
		return err
	}

	d.SetId(resp.Slug)

	return resourceDashboardRead(d, meta)
}

func resourceDashboardRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	slug := d.Id()

	dashboard, err := client.Dashboard(slug)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Printf("[WARN] removing dashboard %s from state because it no longer exists in grafana", slug)
			d.SetId("")
			return nil
		}

		return err
	}

	configJSONBytes, err := json.Marshal(dashboard.Model)
	if err != nil {
		return err
	}

	configJSON := normalizeDashboardConfigJSON(string(configJSONBytes))

	d.SetId(dashboard.Meta.Slug)
	d.Set("slug", dashboard.Meta.Slug)
	d.Set("config_json", configJSON)
	d.Set("folder", dashboard.Folder)

	return nil
}

func resourceDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	dashboard := gapi.Dashboard{}

	dashboard.Model = prepareDashboardModel(d.Get("config_json").(string))

	dashboard.Folder = int64(d.Get("folder").(int))
	dashboard.Overwrite = true

	resp, err := client.NewDashboard(dashboard)
	if err != nil {
		return err
	}

	d.SetId(resp.Slug)

	return resourceDashboardRead(d, meta)
}

func resourceDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	slug := d.Id()
	return client.DeleteDashboard(slug)
}

func prepareDashboardModel(configJSON string) map[string]interface{} {
	configMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(configJSON), &configMap)
	if err != nil {
		// The validate function should've taken care of this.
		panic(fmt.Errorf("Invalid JSON got into prepare func"))
	}

	delete(configMap, "id")
	// Only exists in 5.0+
	delete(configMap, "uid")
	configMap["version"] = 0

	return configMap
}

func validateDashboardConfigJSON(configI interface{}, k string) ([]string, []error) {
	configJSON := configI.(string)
	configMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(configJSON), &configMap)
	if err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

func normalizeDashboardConfigJSON(configI interface{}) string {
	configJSON := configI.(string)

	configMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(configJSON), &configMap)
	if err != nil {
		// The validate function should've taken care of this.
		return ""
	}

	// Some properties are managed by this provider and are thus not
	// significant when included in the JSON.
	delete(configMap, "id")
	delete(configMap, "version")
	// Only exists in 5.0+
	delete(configMap, "uid")

	ret, err := json.Marshal(configMap)
	if err != nil {
		// Should never happen.
		return configJSON
	}

	return string(ret)
}
